/*
Copyright 2017 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package multiplexer

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgproto3/v2"
	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/gravitational/teleport/api/constants"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/auth/native"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/fixtures"
	"github.com/gravitational/teleport/lib/httplib"
	"github.com/gravitational/teleport/lib/jwt"
	"github.com/gravitational/teleport/lib/multiplexer/test"
	"github.com/gravitational/teleport/lib/tlsca"
	"github.com/gravitational/teleport/lib/utils"
	"github.com/gravitational/teleport/lib/utils/cert"
)

func TestMain(m *testing.M) {
	utils.InitLoggerForTests()
	os.Exit(m.Run())
}

// TestMux tests multiplexing protocols
// using the same listener.
func TestMux(t *testing.T) {
	_, signer, err := cert.CreateCertificate("foo", ssh.HostCert)
	require.NoError(t, err)

	// TestMux tests basic use case of multiplexing TLS
	// and SSH on the same listener socket
	t.Run("TLSSSH", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		backend1 := &httptest.Server{
			Listener: mux.TLS(),
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "backend 1")
				}),
			},
		}
		backend1.StartTLS()
		defer backend1.Close()

		go startSSHServer(t, mux.SSH())

		clt, err := ssh.Dial("tcp", listener.Addr().String(), &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second,
		})
		require.NoError(t, err)
		defer clt.Close()

		// Make sure the SSH connection works correctly
		ok, response, err := clt.SendRequest("echo", true, []byte("beep"))
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "beep", string(response))

		client := testClient(backend1)
		re, err := client.Get(backend1.URL)
		require.NoError(t, err)
		defer re.Body.Close()
		bytes, err := io.ReadAll(re.Body)
		require.NoError(t, err)
		require.Equal(t, string(bytes), "backend 1")

		// Close mux, new requests should fail
		mux.Close()
		mux.Wait()

		// use new client to use new connection pool
		client = testClient(backend1)
		re, err = client.Get(backend1.URL)
		if err == nil {
			re.Body.Close()
		}
		require.NotNil(t, err)
	})
	// ProxyLine tests proxy line protocol
	t.Run("ProxyLines", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			desc            string
			proxyLine       []byte
			expectedAddress string
		}{
			{
				desc:            "PROXY protocol v1",
				proxyLine:       []byte(sampleProxyV1Line),
				expectedAddress: "127.0.0.1:12345",
			},
			{
				desc:            "PROXY protocol v2 LOCAL command",
				proxyLine:       sampleProxyV2LineLocal,
				expectedAddress: "", // Shouldn't be changed
			},
			{
				desc:            "PROXY protocol v2 PROXY command",
				proxyLine:       sampleProxyV2Line,
				expectedAddress: "127.0.0.1:12345",
			},
		}

		for _, tt := range testCases {
			t.Run(tt.desc, func(t *testing.T) {
				listener, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)

				mux, err := New(Config{
					Listener:                    listener,
					EnableExternalProxyProtocol: true,
				})
				require.NoError(t, err)
				go mux.Serve()
				defer mux.Close()

				backend1 := &httptest.Server{
					Listener: mux.TLS(),
					Config: &http.Server{
						Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							fmt.Fprintf(w, r.RemoteAddr)
						}),
					},
				}
				backend1.StartTLS()
				defer backend1.Close()

				parsedURL, err := url.Parse(backend1.URL)
				require.NoError(t, err)

				conn, err := net.Dial("tcp", parsedURL.Host)
				require.NoError(t, err)
				defer conn.Close()
				// send proxy line first before establishing TLS connection
				_, err = conn.Write(tt.proxyLine)
				require.NoError(t, err)

				// upgrade connection to TLS
				tlsConn := tls.Client(conn, clientConfig(backend1))
				defer tlsConn.Close()

				// make sure the TLS call succeeded and we got remote address correctly
				out, err := utils.RoundtripWithConn(tlsConn)
				require.NoError(t, err)
				if tt.expectedAddress != "" {
					require.Equal(t, out, tt.expectedAddress)
				} else {
					require.Equal(t, out, tlsConn.LocalAddr().String())
				}
			})
		}
	})

	// TestDisabledProxy makes sure the connection gets dropped
	// when Proxy line support protocol is turned off
	t.Run("DisabledProxy", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: false,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		backend1 := &httptest.Server{
			Listener: mux.TLS(),
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, r.RemoteAddr)
				}),
			},
		}
		backend1.StartTLS()
		defer backend1.Close()

		remoteAddr := net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8000}
		proxyLine := ProxyLine{
			Protocol:    TCP4,
			Source:      remoteAddr,
			Destination: net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000},
		}

		parsedURL, err := url.Parse(backend1.URL)
		require.NoError(t, err)

		conn, err := net.Dial("tcp", parsedURL.Host)
		require.NoError(t, err)
		defer conn.Close()
		// send proxy line first before establishing TLS connection
		_, err = fmt.Fprint(conn, proxyLine.String())
		require.NoError(t, err)

		// upgrade connection to TLS
		tlsConn := tls.Client(conn, clientConfig(backend1))
		defer tlsConn.Close()

		// make sure the TLS call failed
		_, err = utils.RoundtripWithConn(tlsConn)
		require.NotNil(t, err)
	})

	// Timeout test makes sure that multiplexer respects read deadlines.
	t.Run("Timeout", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		config := Config{
			Listener: listener,
			// Set read deadline in the past to remove reliance on real time
			// and simulate scenario when read deadline has elapsed.
			ReadDeadline:                -time.Millisecond,
			EnableExternalProxyProtocol: true,
		}
		mux, err := New(config)
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		backend1 := &httptest.Server{
			Listener: mux.TLS(),
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, r.RemoteAddr)
				}),
			},
		}
		backend1.StartTLS()
		defer backend1.Close()

		parsedURL, err := url.Parse(backend1.URL)
		require.NoError(t, err)

		conn, err := net.Dial("tcp", parsedURL.Host)
		require.NoError(t, err)
		defer conn.Close()

		// upgrade connection to TLS
		tlsConn := tls.Client(conn, clientConfig(backend1))
		defer tlsConn.Close()

		// roundtrip should fail on the timeout
		_, err = utils.RoundtripWithConn(tlsConn)
		require.NotNil(t, err)
	})

	// UnknownProtocol make sure that multiplexer closes connection
	// with unknown protocol
	t.Run("UnknownProtocol", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		conn, err := net.Dial("tcp", listener.Addr().String())
		require.NoError(t, err)
		defer conn.Close()

		// try plain HTTP
		_, err = fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: 127.0.0.1\r\n\r\n")
		require.NoError(t, err)

		// connection should be closed
		_, err = conn.Read(make([]byte, 1))
		require.Equal(t, err, io.EOF)
	})

	// DisableSSH disables SSH
	t.Run("DisableSSH", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		backend1 := &httptest.Server{
			Listener: mux.TLS(),
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "backend 1")
				}),
			},
		}
		backend1.StartTLS()
		defer backend1.Close()

		_, err = ssh.Dial("tcp", listener.Addr().String(), &ssh.ClientConfig{
			Auth:            []ssh.AuthMethod{ssh.Password("abc123")},
			Timeout:         time.Second,
			HostKeyCallback: ssh.FixedHostKey(signer.PublicKey()),
		})
		require.NotNil(t, err)

		// TLS requests will succeed
		client := testClient(backend1)
		re, err := client.Get(backend1.URL)
		require.NoError(t, err)
		defer re.Body.Close()
		bytes, err := io.ReadAll(re.Body)
		require.NoError(t, err)
		require.Equal(t, string(bytes), "backend 1")

		// Close mux, new requests should fail
		mux.Close()
		mux.Wait()

		// use new client to use new connection pool
		client = testClient(backend1)
		re, err = client.Get(backend1.URL)
		if err == nil {
			re.Body.Close()
		}
		require.NotNil(t, err)
	})

	// TestDisableTLS tests scenario with disabled TLS
	t.Run("DisableTLS", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		backend1 := &httptest.Server{
			Listener: &noopListener{addr: listener.Addr()},
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, "backend 1")
				}),
			},
		}
		backend1.StartTLS()
		defer backend1.Close()

		go startSSHServer(t, mux.SSH())

		clt, err := ssh.Dial("tcp", listener.Addr().String(), &ssh.ClientConfig{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         time.Second,
		})
		require.NoError(t, err)
		defer clt.Close()

		// Make sure the SSH connection works correctly
		ok, response, err := clt.SendRequest("echo", true, []byte("beep"))
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "beep", string(response))

		client := testClient(backend1)
		re, err := client.Get(backend1.URL)
		if err == nil {
			re.Body.Close()
		}
		require.NotNil(t, err)

		// Close mux, new requests should fail
		mux.Close()
		mux.Wait()
	})

	// NextProto tests multiplexing using NextProto selector
	t.Run("NextProto", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		cfg, err := fixtures.LocalTLSConfig()
		require.NoError(t, err)

		tlsLis, err := NewTLSListener(TLSListenerConfig{
			Listener: tls.NewListener(mux.TLS(), cfg.TLS),
		})
		require.NoError(t, err)
		go tlsLis.Serve()

		opts := []grpc.ServerOption{
			grpc.Creds(&httplib.TLSCreds{
				Config: cfg.TLS,
			}),
		}
		s := grpc.NewServer(opts...)
		test.RegisterPingerServer(s, &server{})

		errCh := make(chan error, 2)

		go func() {
			errCh <- s.Serve(tlsLis.HTTP2())
		}()

		httpServer := http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "http backend")
			}),
		}
		go func() {
			err := httpServer.Serve(tlsLis.HTTP())
			if err == nil || err == http.ErrServerClosed {
				errCh <- nil
				return
			}
			errCh <- err
		}()

		url := fmt.Sprintf("https://%s", listener.Addr())
		client := cfg.NewClient()
		re, err := client.Get(url)
		require.NoError(t, err)
		defer re.Body.Close()
		bytes, err := io.ReadAll(re.Body)
		require.NoError(t, err)
		require.Equal(t, string(bytes), "http backend")

		creds := credentials.NewClientTLSFromCert(cfg.CertPool, "")

		// Set up a connection to the server.
		conn, err := grpc.Dial(listener.Addr().String(), grpc.WithTransportCredentials(creds), grpc.WithBlock())
		require.NoError(t, err)
		defer conn.Close()

		gclient := test.NewPingerClient(conn)

		out, err := gclient.Ping(context.TODO(), &test.Request{})
		require.NoError(t, err)
		require.Equal(t, out.GetPayload(), "grpc backend")

		// Close mux, new requests should fail
		mux.Close()
		mux.Wait()

		// use new client to use new connection pool
		client = cfg.NewClient()
		re, err = client.Get(url)
		if err == nil {
			re.Body.Close()
		}
		require.NotNil(t, err)

		httpServer.Close()
		s.Stop()
		// wait for both servers to finish
		for i := 0; i < 2; i++ {
			err := <-errCh
			require.NoError(t, err)
		}
	})

	t.Run("PostgresProxy", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		mux, err := New(Config{
			Context:  ctx,
			Listener: listener,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		// register listener before establishing frontend connection
		dblistener := mux.DB()

		// Connect to the listener and send Postgres SSLRequest which is what
		// psql or other Postgres client will do.
		conn, err := net.Dial("tcp", listener.Addr().String())
		require.NoError(t, err)
		defer conn.Close()

		frontend := pgproto3.NewFrontend(pgproto3.NewChunkReader(conn), conn)
		err = frontend.Send(&pgproto3.SSLRequest{})
		require.NoError(t, err)

		// This should not hang indefinitely since we set timeout on the mux context above.
		conn, err = dblistener.Accept()
		require.NoError(t, err, "detected Postgres connection")
		require.Equal(t, ProtoPostgres, conn.(*Conn).Protocol())
	})

	// WebListener verifies web listener correctly multiplexes connections
	// between web and database listeners based on the client certificate.
	t.Run("WebListener", func(t *testing.T) {
		t.Parallel()
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
		})
		require.NoError(t, err)
		go mux.Serve()
		defer mux.Close()

		// register listener before establishing frontend connection
		tlslistener := mux.TLS()

		// Generate self-signed CA.
		caKey, caCert, err := tlsca.GenerateSelfSignedCA(pkix.Name{CommonName: "test-ca"}, nil, time.Hour)
		require.NoError(t, err)
		ca, err := tlsca.FromKeys(caCert, caKey)
		require.NoError(t, err)
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM(caCert)

		// Sign server certificate.
		serverRSAKey, err := native.GenerateRSAPrivateKey()
		require.NoError(t, err)
		serverPEM, err := ca.GenerateCertificate(tlsca.CertificateRequest{
			Subject:   pkix.Name{CommonName: "localhost"},
			PublicKey: serverRSAKey.Public(),
			NotAfter:  time.Now().Add(time.Hour),
			DNSNames:  []string{"127.0.0.1"},
		})
		require.NoError(t, err)
		serverCert, err := tls.X509KeyPair(serverPEM, tlsca.MarshalPrivateKeyPEM(serverRSAKey))
		require.NoError(t, err)

		// Sign client certificate with database access identity.
		clientRSAKey, err := rsa.GenerateKey(rand.Reader, constants.RSAKeySize)
		require.NoError(t, err)
		subject, err := (&tlsca.Identity{
			Username: "alice",
			Groups:   []string{"admin"},
			RouteToDatabase: tlsca.RouteToDatabase{
				ServiceName: "postgres",
			},
		}).Subject()
		require.NoError(t, err)
		clientPEM, err := ca.GenerateCertificate(tlsca.CertificateRequest{
			Subject:   subject,
			PublicKey: clientRSAKey.Public(),
			NotAfter:  time.Now().Add(time.Hour),
		})
		require.NoError(t, err)
		clientCert, err := tls.X509KeyPair(clientPEM, tlsca.MarshalPrivateKeyPEM(clientRSAKey))
		require.NoError(t, err)

		webLis, err := NewWebListener(WebListenerConfig{
			Listener: tls.NewListener(tlslistener, &tls.Config{
				ClientCAs:    certPool,
				ClientAuth:   tls.VerifyClientCertIfGiven,
				Certificates: []tls.Certificate{serverCert},
			}),
		})
		require.NoError(t, err)
		go webLis.Serve()
		defer webLis.Close()

		go func() {
			conn, err := webLis.Web().Accept()
			require.NoError(t, err)
			defer conn.Close()
			conn.Write([]byte("web listener"))
		}()

		go func() {
			conn, err := webLis.DB().Accept()
			require.NoError(t, err)
			defer conn.Close()
			conn.Write([]byte("db listener"))
		}()

		webConn, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{
			RootCAs: certPool,
		})
		require.NoError(t, err)
		defer webConn.Close()

		webBytes, err := io.ReadAll(webConn)
		require.NoError(t, err)
		require.Equal(t, "web listener", string(webBytes))

		dbConn, err := tls.Dial("tcp", listener.Addr().String(), &tls.Config{
			RootCAs:      certPool,
			Certificates: []tls.Certificate{clientCert},
		})
		require.NoError(t, err)
		defer dbConn.Close()

		dbBytes, err := io.ReadAll(dbConn)
		require.NoError(t, err)
		require.Equal(t, "db listener", string(dbBytes))
	})

	// Ensures that we can correctly send and verify signed PROXY header
	t.Run("signed PROXYv2 headers", func(t *testing.T) {
		t.Parallel()

		const clusterName = "teleport-test"
		tlsProxyCert, casGetter, jwtSigner := getTestCertCAsGetterAndSigner(t, clusterName)

		listener4, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)

		// If listener for IPv6 will fail to be created we'll skip IPv6 portion of test.
		listener6, _ := net.Listen("tcp6", "[::1]:0")

		startServing := func(muxListener net.Listener) (*Mux, *httptest.Server) {
			mux, err := New(Config{
				Listener:                    muxListener,
				EnableExternalProxyProtocol: true,
				CertAuthorityGetter:         casGetter,
				Clock:                       clockwork.NewFakeClockAt(time.Now()),
				LocalClusterName:            clusterName,
			})
			require.NoError(t, err)

			muxTLSListener := mux.TLS()

			go mux.Serve()

			backend := &httptest.Server{
				Listener: muxTLSListener,

				Config: &http.Server{
					Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintf(w, r.RemoteAddr)
					}),
				},
			}
			backend.StartTLS()

			return mux, backend
		}

		mux4, backend4 := startServing(listener4)
		defer mux4.Close()
		defer backend4.Close()

		var backend6 *httptest.Server
		var mux6 *Mux
		if listener6 != nil {
			mux6, backend6 = startServing(listener6)
			defer mux6.Close()
			defer backend6.Close()
		}

		addr1 := net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 444}
		addr2 := net.TCPAddr{IP: net.ParseIP("5.4.3.2"), Port: 555}
		addrV6 := net.TCPAddr{IP: net.ParseIP("::1"), Port: 999}

		t.Run("single signed PROXY header", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addr1, &addr2, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			_, err = conn.Write(signedHeader)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			out, err := utils.RoundtripWithConn(clt)
			require.NoError(t, err)
			require.Equal(t, addr1.String(), out)
		})
		t.Run("single signed PROXY header on IPv6", func(t *testing.T) {
			if listener6 == nil {
				t.Skip("Skipping since IPv6 listener is not available")
			}
			conn, err := net.Dial("tcp6", listener6.Addr().String())
			require.NoError(t, err)

			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addrV6, &addrV6, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			_, err = conn.Write(signedHeader)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend6))

			out, err := utils.RoundtripWithConn(clt)
			require.NoError(t, err)
			require.Equal(t, addrV6.String(), out)
		})
		t.Run("two signed PROXY headers", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addr1, &addr2, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			_, err = conn.Write(signedHeader)
			require.NoError(t, err)
			_, err = conn.Write(signedHeader)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			_, err = utils.RoundtripWithConn(clt)
			require.Error(t, err)
		})
		t.Run("two signed PROXY headers, one signed for wrong cluster", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addr1, &addr2, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)
			signedHeader2, err := signPROXYHeader(&addr2, &addr1, clusterName+"wrong", tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			_, err = conn.Write(signedHeader)
			require.NoError(t, err)
			_, err = conn.Write(signedHeader2)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			_, err = utils.RoundtripWithConn(clt)
			require.Error(t, err)
		})
		t.Run("first unsigned then signed PROXY headers", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addr1, &addr2, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			pl := ProxyLine{
				Protocol:    TCP4,
				Source:      addr2,
				Destination: addr1,
			}

			b, err := pl.Bytes()
			require.NoError(t, err)

			_, err = conn.Write(b)
			require.NoError(t, err)
			_, err = conn.Write(signedHeader)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			out, err := utils.RoundtripWithConn(clt)
			require.NoError(t, err)
			require.Equal(t, addr1.String(), out)
		})
		t.Run("first signed then unsigned PROXY headers", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			signedHeader, err := signPROXYHeader(&addr1, &addr2, clusterName, tlsProxyCert, jwtSigner)
			require.NoError(t, err)

			pl := ProxyLine{
				Protocol:    TCP4,
				Source:      addr2,
				Destination: addr1,
			}

			b, err := pl.Bytes()
			require.NoError(t, err)

			_, err = conn.Write(signedHeader)
			require.NoError(t, err)
			_, err = conn.Write(b)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			out, err := utils.RoundtripWithConn(clt)
			require.NoError(t, err)
			require.Equal(t, addr1.String(), out)
		})
		t.Run("two unsigned PROXY headers, gets an error", func(t *testing.T) {
			conn, err := net.Dial("tcp", listener4.Addr().String())
			require.NoError(t, err)
			defer conn.Close()

			pl := ProxyLine{
				Protocol:    TCP4,
				Source:      addr2,
				Destination: addr1,
			}

			b, err := pl.Bytes()
			require.NoError(t, err)

			_, err = conn.Write(b)
			require.NoError(t, err)
			_, err = conn.Write(b)
			require.NoError(t, err)

			clt := tls.Client(conn, clientConfig(backend4))

			_, err = utils.RoundtripWithConn(clt)
			require.Error(t, err)
		})
	})
	// Ensures that we can correctly send and verify signed PROXY header
	t.Run("signed PROXY header is ignored if signed by wrong cluster", func(t *testing.T) {
		t.Parallel()

		const clusterName = "teleport-test"
		tlsProxyCert, _, jwtSigner := getTestCertCAsGetterAndSigner(t, clusterName)
		_, wrongCAsGetter, _ := getTestCertCAsGetterAndSigner(t, "wrong-cluster")

		listener, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)

		mux, err := New(Config{
			Listener:                    listener,
			EnableExternalProxyProtocol: true,
			CertAuthorityGetter:         wrongCAsGetter,
			LocalClusterName:            "different-cluster",
		})
		require.NoError(t, err)

		muxTLSListener := mux.TLS()

		go mux.Serve()
		defer mux.Close()

		backend := &httptest.Server{
			Listener: muxTLSListener,
			Config: &http.Server{
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, r.RemoteAddr)
				}),
			},
		}
		backend.StartTLS()
		defer backend.Close()

		conn, err := net.Dial("tcp", listener.Addr().String())
		require.NoError(t, err)
		defer conn.Close()

		ip := "1.2.3.4"
		sAddr := net.TCPAddr{IP: net.ParseIP(ip), Port: 444}
		dAddr := net.TCPAddr{IP: net.ParseIP(ip), Port: 555}

		signedHeader, err := signPROXYHeader(&sAddr, &dAddr, clusterName, tlsProxyCert, jwtSigner)
		require.NoError(t, err)

		_, err = conn.Write(signedHeader)
		require.NoError(t, err)

		clt := tls.Client(conn, clientConfig(backend))

		out, err := utils.RoundtripWithConn(clt)
		require.NoError(t, err)
		require.Equal(t, conn.LocalAddr().String(), out)
	})
}

func TestProtocolString(t *testing.T) {
	for i := -1; i < len(protocolStrings)+1; i++ {
		got := Protocol(i).String()
		switch i {
		case -1, len(protocolStrings) + 1:
			require.Equal(t, "", got)
		default:
			require.Equal(t, protocolStrings[Protocol(i)], got)
		}
	}
}

// server is used to implement test.PingerServer
type server struct {
	test.UnimplementedPingerServer
}

func (s *server) Ping(ctx context.Context, req *test.Request) (*test.Response, error) {
	return &test.Response{Payload: "grpc backend"}, nil
}

// clientConfig returns tls client config from test http server
// set up to listen on TLS
func clientConfig(srv *httptest.Server) *tls.Config {
	cert, err := x509.ParseCertificate(srv.TLS.Certificates[0].Certificate[0])
	if err != nil {
		panic(err)
	}

	certpool := x509.NewCertPool()
	certpool.AddCert(cert)
	return &tls.Config{
		RootCAs:    certpool,
		ServerName: fmt.Sprintf("%v", cert.IPAddresses[0].String()),
	}
}

// testClient is a test HTTP client set up for TLS
func testClient(srv *httptest.Server) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: clientConfig(srv),
		},
	}
}

type noopListener struct {
	addr net.Addr
}

func (noopListener) Accept() (net.Conn, error) {
	return nil, errors.New("noop")
}

func (noopListener) Close() error {
	return nil
}

func (l noopListener) Addr() net.Addr {
	return l.addr
}

func TestIsHTTP(t *testing.T) {
	t.Parallel()
	for _, verb := range httpMethods {
		t.Run(fmt.Sprintf("Accept %v", string(verb)), func(t *testing.T) {
			data := fmt.Sprintf("%v /some/path HTTP/1.1", string(verb))
			require.True(t, isHTTP([]byte(data)))
		})
	}

	rejectedInputs := []string{
		"some random junk",
		"FAKE /some/path HTTP/1.1",
		// This case checks for a bug where the arguments to bytes.HasPrefix are reversed.
		"GE",
	}
	for _, input := range rejectedInputs {
		t.Run(fmt.Sprintf("Reject %q", input), func(t *testing.T) {
			require.False(t, isHTTP([]byte(input)))
		})
	}
}

func getTestCertCAsGetterAndSigner(t testing.TB, clusterName string) ([]byte, CertAuthorityGetter, JWTPROXYSigner) {
	t.Helper()
	caPriv, caCert, err := tlsca.GenerateSelfSignedCA(pkix.Name{
		CommonName: clusterName, Organization: []string{clusterName}}, []string{clusterName}, time.Hour)
	require.NoError(t, err)

	tlsCA, err := tlsca.FromKeys(caCert, caPriv)
	require.NoError(t, err)

	ca, err := types.NewCertAuthority(types.CertAuthoritySpecV2{
		Type:        types.HostCA,
		ClusterName: clusterName,
		ActiveKeys: types.CAKeySet{
			TLS: []*types.TLSKeyPair{
				{
					Cert: caCert,
					Key:  caPriv,
				},
			},
		},
	})
	require.NoError(t, err)

	mockCAGetter := func(ctx context.Context, id types.CertAuthID, loadKeys bool) (types.CertAuthority, error) {
		return ca, nil
	}
	proxyPriv, err := rsa.GenerateKey(rand.Reader, constants.RSAKeySize)
	require.NoError(t, err)

	// Create host identity with role "Proxy"
	identity := tlsca.Identity{
		TeleportCluster: clusterName,
		Username:        "proxy1",
		Groups:          []string{string(types.RoleProxy)},
		Expires:         time.Now().Add(time.Hour),
	}

	subject, err := identity.Subject()
	require.NoError(t, err)
	certReq := tlsca.CertificateRequest{
		PublicKey: proxyPriv.Public(),
		Subject:   subject,
		NotAfter:  time.Now().Add(time.Hour),
		DNSNames:  []string{"localhost", "127.0.0.1"},
	}
	tlsProxyCertPEM, err := tlsCA.GenerateCertificate(certReq)
	require.NoError(t, err)
	clock := clockwork.NewFakeClockAt(time.Now())
	jwtSigner, err := jwt.New(&jwt.Config{
		Clock:       clock,
		Algorithm:   defaults.ApplicationTokenAlgorithm,
		ClusterName: clusterName,
		PrivateKey:  proxyPriv,
	})
	require.NoError(t, err)

	tlsProxyCertDER, err := tlsca.ParseCertificatePEM(tlsProxyCertPEM)
	require.NoError(t, err)

	return tlsProxyCertDER.Raw, mockCAGetter, jwtSigner
}

func startSSHServer(t *testing.T, listener net.Listener) {
	nConn, err := listener.Accept()
	assert.NoError(t, err)

	t.Cleanup(func() { nConn.Close() })

	block, _ := pem.Decode(fixtures.LocalhostKey)
	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	assert.NoError(t, err)

	signer, err := ssh.NewSignerFromKey(pkey)
	assert.NoError(t, err)

	config := &ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(signer)

	conn, _, reqs, err := ssh.NewServerConn(nConn, config)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	t.Cleanup(func() { conn.Close() })

	go func() {
		for newReq := range reqs {
			if newReq.Type == "echo" {
				newReq.Reply(true, newReq.Payload)
			}
			err := newReq.Reply(false, nil)
			assert.NoError(t, err)
		}
	}()
}

func BenchmarkMux_ProxyV2Signature(b *testing.B) {
	const clusterName = "test-teleport"

	clock := clockwork.NewFakeClockAt(time.Now())
	tlsProxyCert, caGetter, jwtSigner := getTestCertCAsGetterAndSigner(b, clusterName)

	ca, err := caGetter(context.Background(), types.CertAuthID{
		Type:       types.HostCA,
		DomainName: clusterName,
	}, false)
	require.NoError(b, err)

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(ca.GetTrustedTLSKeyPairs()[0].Cert)
	require.True(b, ok)

	ip := "1.2.3.4"
	sAddr := net.TCPAddr{IP: net.ParseIP(ip), Port: 444}
	dAddr := net.TCPAddr{IP: net.ParseIP(ip), Port: 555}

	b.Run("simulation of signing and verifying PROXY header", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			token, err := jwtSigner.SignPROXYJWT(jwt.PROXYSignParams{
				ClusterName:        clusterName,
				SourceAddress:      sAddr.String(),
				DestinationAddress: dAddr.String(),
			})
			require.NoError(b, err)

			pl := ProxyLine{
				Protocol:    TCP4,
				Source:      sAddr,
				Destination: dAddr,
			}
			err = pl.AddSignature([]byte(token), tlsProxyCert)
			require.NoError(b, err)

			_, err = pl.Bytes()
			require.NoError(b, err)

			cert, err := tlsca.ParseCertificatePEM(tlsProxyCert)
			require.NoError(b, err)
			chains, err := cert.Verify(x509.VerifyOptions{Roots: roots})
			require.NoError(b, err)
			require.NotNil(b, chains)

			identity, err := tlsca.FromSubject(cert.Subject, cert.NotAfter)
			require.NoError(b, err)

			foundRole := checkForSystemRole(identity, types.RoleProxy)
			require.True(b, foundRole, "Missing 'Proxy' role on the signing certificate")

			// Check JWT using proxy cert's public key
			jwtVerifier, err := jwt.New(&jwt.Config{
				Clock:       clock,
				PublicKey:   cert.PublicKey,
				Algorithm:   defaults.ApplicationTokenAlgorithm,
				ClusterName: clusterName,
			})
			require.NoError(b, err, "Could not create JWT verifier")

			claims, err := jwtVerifier.VerifyPROXY(jwt.PROXYVerifyParams{
				ClusterName:        clusterName,
				SourceAddress:      sAddr.String(),
				DestinationAddress: dAddr.String(),
				RawToken:           token,
			})
			require.NoError(b, err, "Got an error while verifying PROXY JWT")
			require.NotNil(b, claims)
			require.Equal(b, fmt.Sprintf("%s/%s", sAddr.String(), dAddr.String()), claims.Subject,
				"IP addresses in PROXY header don't match JWT")
		}
	})
}

func Test_GetTcpAddr(t *testing.T) {
	testCases := []struct {
		input    net.Addr
		expected string
	}{
		{
			input: &utils.NetAddr{
				Addr:        "127.0.0.1:24998",
				AddrNetwork: "tcp",
				Path:        "",
			},
			expected: "127.0.0.1:24998",
		},
		{
			input:    nil,
			expected: ":0",
		},
		{
			input: &net.TCPAddr{
				IP:   net.ParseIP("8.8.8.8"),
				Port: 25000,
			},
			expected: "8.8.8.8:25000",
		},
		{
			input: &net.TCPAddr{
				IP:   net.ParseIP("::1"),
				Port: 25000,
			},
			expected: "[::1]:25000",
		},
	}

	for _, tt := range testCases {
		result := getTCPAddr(tt.input)
		require.Equal(t, tt.expected, result.String())
	}
}

func TestIsDifferentTCPVersion(t *testing.T) {
	testCases := []struct {
		addr1    string
		addr2    string
		expected bool
	}{
		{
			addr1:    "8.8.8.8:42",
			addr2:    "8.8.8.8:42",
			expected: false,
		},
		{
			addr1:    "[2601:602:8700:4470:a3:813c:1d8c:30b9]:42",
			addr2:    "[2607:f8b0:4005:80a::200e]:42",
			expected: false,
		},
		{
			addr1:    "127.0.0.1:42",
			addr2:    "[::1]:42",
			expected: true,
		},
		{
			addr1:    "[::1]:42",
			addr2:    "127.0.0.1:42",
			expected: true,
		},
		{
			addr1:    "::ffff:39.156.68.48:42",
			addr2:    "39.156.68.48:42",
			expected: true,
		},
		{
			addr1:    "[2607:f8b0:4005:80a::200e]:42",
			addr2:    "1.1.1.1:42",
			expected: true,
		},
		{
			addr1:    "127.0.0.1:42",
			addr2:    "[2607:f8b0:4005:80a::200e]:42",
			expected: true,
		},
		{
			addr1:    "::ffff:39.156.68.48:42",
			addr2:    "[2607:f8b0:4005:80a::200e]:42",
			expected: false,
		},
	}

	for _, tt := range testCases {
		addr1 := getTCPAddr(utils.MustParseAddr(tt.addr1))
		addr2 := getTCPAddr(utils.MustParseAddr(tt.addr2))
		require.Equal(t, tt.expected, isDifferentTCPVersion(addr1, addr2),
			fmt.Sprintf("Unexpected result for %q, %q", tt.addr1, tt.addr2))
	}
}
