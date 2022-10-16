package msgbroker_nsq

type NSQMessage struct {
	Data *[]byte
}

func (m *NSQMessage) GetData() *[]byte {
	return m.Data
}
