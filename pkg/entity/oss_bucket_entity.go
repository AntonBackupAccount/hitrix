package entity

import "github.com/latolukasz/beeorm/v2"

type OSSBucketCounterEntity struct {
	beeorm.ORM `orm:"table=oss_buckets_counters"`
	ID         uint64
	Counter    uint64 `orm:"required"`
}
