/*
   Copyright 2022 Gravitational, Inc.

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

package web

import (
	"net/http"
	"os"

	"github.com/gravitational/trace"
	"github.com/julienschmidt/httprouter"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/paymentintent"
)

// createPaymentIntentHandle sends a user event to the UserEvent service
// this handler is for on-boarding user events pre-session
func (h *Handler) createPaymentIntentHandle(w http.ResponseWriter, r *http.Request, params httprouter.Params, sctx *SessionContext) (interface{}, error) {
	// todo mberg move the logic to a client
	//   actually can we use grpc?
	//client := h.cfg.ProxyClient
	//response, err := client.GetPaymentIntent(r.Context())

	stripe.Key = os.Getenv("STRIPE_SK")

	// todo mberg set params dynamically
	sParams := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(1099),
		Currency: stripe.String(string(stripe.CurrencyUSD)),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}
	response, err := paymentintent.New(sParams)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	data := CheckoutData{
		ClientSecret: response.ClientSecret,
	}

	return data, nil
}

type CheckoutData struct {
	ClientSecret string `json:"clientSecret"`
}
