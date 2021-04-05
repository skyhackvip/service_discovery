package configs

import (
	"time"
)

const (
	NodeStatusUp = iota
	NodeStatusDown
)

const (
	RegisterURL = "/api/register"
	CancelURL   = "/api/cancel"
	RenewURL    = "/api/renew"
	FetchAllURL = "/api/fetchall"
)

const (
	DiscoveryAppId = "Kavin.discovery"
)

const (
	RenewInterval               = 30 * time.Second   //client heart beat interval
	CheckEvictInterval          = 60 * time.Second   //evict task interval
	SelfProtectThreshold        = 0.85               //self protect threshold
	ResetGuardNeedCountInterval = 15 * time.Minute   //ticker reset guard need count
	InstanceExpireDuration      = 90 * time.Second   //instance's renewTimestamp after this will be canceled
	InstanceMaxExpireDuration   = 3600 * time.Second //instance's renewTimestamp after this will be canceled
)
