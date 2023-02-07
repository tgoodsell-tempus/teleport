/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// package: teleport.terminal.v1
// file: v1/gateway.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class Gateway extends jspb.Message { 
    getUri(): string;
    setUri(value: string): Gateway;

    getTargetName(): string;
    setTargetName(value: string): Gateway;

    getTargetUri(): string;
    setTargetUri(value: string): Gateway;

    getTargetUser(): string;
    setTargetUser(value: string): Gateway;

    getLocalAddress(): string;
    setLocalAddress(value: string): Gateway;

    getLocalPort(): string;
    setLocalPort(value: string): Gateway;

    getProtocol(): string;
    setProtocol(value: string): Gateway;

    getCliCommand(): string;
    setCliCommand(value: string): Gateway;

    getTargetSubresourceName(): string;
    setTargetSubresourceName(value: string): Gateway;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Gateway.AsObject;
    static toObject(includeInstance: boolean, msg: Gateway): Gateway.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Gateway, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Gateway;
    static deserializeBinaryFromReader(message: Gateway, reader: jspb.BinaryReader): Gateway;
}

export namespace Gateway {
    export type AsObject = {
        uri: string,
        targetName: string,
        targetUri: string,
        targetUser: string,
        localAddress: string,
        localPort: string,
        protocol: string,
        cliCommand: string,
        targetSubresourceName: string,
    }
}
