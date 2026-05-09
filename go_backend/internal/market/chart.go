package market

import (
	"time"

	"iwx/go_backend/internal/domain"
)

var (
	defaultLookbackSeconds int64 = 24 * 60 * 60
	defaultBucketSeconds   int64 = 5 * 60
	supportedLookbacks           = map[int64]struct{}{
		30 * 24 * 60 * 60: {},
		5 * 24 * 60 * 60:  {},
		24 * 60 * 60:      {},
	}
	supportedBuckets = map[int64]struct{}{
		5 * 60:       {},
		60 * 60:      {},
		24 * 60 * 60: {},
	}
)

func DefaultChartConfig() domain.ChartConfig {
	return domain.ChartConfig{
		LookbackSeconds: defaultLookbackSeconds,
		BucketSeconds:   defaultBucketSeconds,
	}
}

func NormalizeChartConfig(lookbackSeconds, bucketSeconds *int64) domain.ChartConfig {
	config := DefaultChartConfig()

	if lookbackSeconds != nil {
		if _, ok := supportedLookbacks[*lookbackSeconds]; ok {
			config.LookbackSeconds = *lookbackSeconds
		}
	}

	if bucketSeconds != nil {
		if _, ok := supportedBuckets[*bucketSeconds]; ok {
			config.BucketSeconds = *bucketSeconds
		}
	}

	return config
}

func BucketStart(timestamp time.Time, bucketSeconds int64) time.Time {
	if timestamp.IsZero() {
		return time.Time{}
	}

	unix := timestamp.UTC().Unix()
	return time.Unix((unix/bucketSeconds)*bucketSeconds, 0).UTC()
}

func BucketSnapshots(
	snapshots []domain.MarketSnapshot,
	bucketSeconds int64,
	firstBucket time.Time,
	lastBucket time.Time,
) []domain.ChartPoint {
	grouped := map[time.Time][]domain.MarketSnapshot{}
	for _, snapshot := range snapshots {
		bucket := BucketStart(snapshot.InsertedAt, bucketSeconds)
		grouped[bucket] = append(grouped[bucket], snapshot)
	}

	rows := []domain.ChartPoint{}
	carry := domain.ChartPoint{}
	for bucket := firstBucket; !bucket.After(lastBucket); bucket = bucket.Add(time.Duration(bucketSeconds) * time.Second) {
		bucketSnapshot := reduceBucketSnapshots(grouped[bucket], bucket)
		carry = carryForward(carry, bucketSnapshot)

		row := domain.ChartPoint{
			BucketStart: bucket,
			InsertedAt:  bucket,
			MidAbove:    carry.MidAbove,
			MidBelow:    carry.MidBelow,
			BestAbove:   carry.BestAbove,
			BestBelow:   carry.BestBelow,
		}

		if !carry.InsertedAt.IsZero() {
			row.InsertedAt = carry.InsertedAt
		}

		rows = append(rows, row)
	}

	return rows
}

func reduceBucketSnapshots(snapshots []domain.MarketSnapshot, bucket time.Time) domain.ChartPoint {
	point := domain.ChartPoint{BucketStart: bucket}
	for _, snapshot := range snapshots {
		point.InsertedAt = snapshot.InsertedAt
		if snapshot.MidAbove != nil {
			point.MidAbove = snapshot.MidAbove
		}
		if snapshot.MidBelow != nil {
			point.MidBelow = snapshot.MidBelow
		}
		if snapshot.BestAbove != nil {
			point.BestAbove = snapshot.BestAbove
		}
		if snapshot.BestBelow != nil {
			point.BestBelow = snapshot.BestBelow
		}
	}

	return point
}

func carryForward(carry, next domain.ChartPoint) domain.ChartPoint {
	if !next.InsertedAt.IsZero() {
		carry.InsertedAt = next.InsertedAt
	}
	if next.MidAbove != nil {
		carry.MidAbove = next.MidAbove
	}
	if next.MidBelow != nil {
		carry.MidBelow = next.MidBelow
	}
	if next.BestAbove != nil {
		carry.BestAbove = next.BestAbove
	}
	if next.BestBelow != nil {
		carry.BestBelow = next.BestBelow
	}

	return carry
}
