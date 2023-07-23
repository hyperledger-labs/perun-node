// Copyright (c) 2020 - for information on the respective copyright owner
// see the NOTICE file and/or the repository at
// https://github.com/hyperledger-labs/perun-node
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package perun

//go:generate protoc --proto_path=proto --go_out=api/grpc/pb --go-grpc_out=api/grpc/pb proto/nodetypes.proto proto/errors.proto proto/payment_service.proto
//go:generate protoc --proto_path=proto --go_out=api/grpc/pb --go-grpc_out=api/grpc/pb proto/sdktypes.proto proto/funding_service.proto
