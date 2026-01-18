package mqtt

import (
	"log"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client MQTT.Client
}

func New(o *MQTT.ClientOptions) (*MQTTClient, error) {
	o.OnConnect = func(c MQTT.Client) {
		log.Println("mqtt connected")
	}

	o.OnConnectionLost = func(c MQTT.Client, err error) {
		log.Printf("mqtt connection lost: %v", err)
	}

	o.OnReconnecting = func(c MQTT.Client, co *MQTT.ClientOptions) {
		log.Printf("mqtt reconnecting to %v", co.Servers)
	}

	client := MQTT.NewClient(o)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	return &MQTTClient{
		client: client,
	}, nil
}

func (m *MQTTClient) Publish(topic string, qos byte, retain bool, payload []byte) error {
	token := m.client.Publish(topic, qos, retain, payload)
	token.Wait()
	return token.Error()
}

func (m *MQTTClient) Close() {
	m.client.Disconnect(250)
}
