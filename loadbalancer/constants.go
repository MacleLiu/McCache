package loadbalancer

import "time"

const (
	Policy             = "consistent_hash_policy"
	Key                = "data_key"
	connectionLifetime = time.Second * 5
)
