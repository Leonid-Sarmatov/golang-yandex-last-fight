package heartbeat

import (

)

type Heartbeat struct {

}

func NewHeartbeat() *Heartbeat {
	var heartbeat Heartbeat
	return &heartbeat
}

func (h *Heartbeat) Ping() error {
	return nil
}