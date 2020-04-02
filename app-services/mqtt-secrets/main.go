//
// Copyright (c) 2020 Intel Corporation
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"fmt"
	"os"

	"github.com/edgexfoundry/app-functions-sdk-go/appsdk"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/transforms"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	serviceKey = "mqtt-secrets"
)

var counter int

func main() {

	// 1) First thing to do is to create an instance of the EdgeX SDK and initialize it.
	edgexSdk := &appsdk.AppFunctionsSDK{ServiceKey: serviceKey}
	if err := edgexSdk.Initialize(); err != nil {
		message := fmt.Sprintf("SDK initialization failed: %v\n", err)
		if edgexSdk.LoggingClient != nil {
			edgexSdk.LoggingClient.Error(message)
		} else {
			fmt.Println(message)
		}
		os.Exit(-1)
	}

	// 2) shows how to access the application's specific configuration settings.
	deviceNames, err := edgexSdk.GetAppSettingStrings("DeviceNames")
	if err != nil {
		edgexSdk.LoggingClient.Error(err.Error())
		os.Exit(-1)
	}
	edgexSdk.LoggingClient.Info(fmt.Sprintf("Filtering for devices %v", deviceNames))

	// Since we are using MQTT, we'll also need to set up the addressable model to
	// configure it to send to our broker. If you don't have a broker setup you can pull one from docker i.e:
	// docker run -it -p 1883:1883 -p 9001:9001  eclipse-mosquitto
	addressable := models.Addressable{
		Address:   "localhost",
		Port:      8883,
		Protocol:  "tls",
		Publisher: "clientid",
		User:      "",
		Password:  "",
		Topic:     "sampleTopic",
	}

	// Using default settings, so not changing any fields in MqttConfig
	mqttConfig := transforms.MqttConfig{
		SkipCertVerify: true,
	}

	// Make sure you change KeyFile and CertFile here to point to actual key/cert files
	// or an error will be logged for failing to load key/cert files
	// If you don't use key/cert for MQTT authentication, just pass nil to NewMQTTSender() as following:
	// mqttSender := transforms.NewMQTTSender(edgexSdk.LoggingClient, addressable, nil, mqttConfig)
	// pair := transforms.KeyCertPair{
	// 	KeyFile:  "PATH_TO_YOUR_KEY_FILE",
	// 	CertFile: "PATH_TO_YOUR_CERT_FILE",
	// }

	// 3) This is our pipeline configuration, the collection of functions to
	// execute every time an event is triggered.
	edgexSdk.SetFunctionsPipeline(
		transforms.NewMQTTSecretSender(edgexSdk.LoggingClient, addressable, "/mqtt", mqttConfig, false).MQTTSend,
	)

	// 4) Lastly, we'll go ahead and tell the SDK to "start" and begin listening for events
	// to trigger the pipeline.
	err = edgexSdk.MakeItRun()
	if err != nil {
		edgexSdk.LoggingClient.Error("MakeItRun returned error: ", err.Error())
		os.Exit(-1)
	}

	// Do any required cleanup here

	os.Exit(0)
}
