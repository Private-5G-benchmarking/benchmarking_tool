package samplelib

import (
	"math/rand"
)

func Sample(cdf []float32) int {
	r := rand.Float32()

	bucket := 0

	for r > cdf[bucket] {
		bucket++
	}
	return bucket
}

func GetBinaryCdf(sample_prob float32) []float32 {
	pdf := []float32{1 - sample_prob, sample_prob}
	cdf := []float32{0.0, 0.0}
	cdf[0] = pdf[0]

	for i := 1; i < 2; i++ {
		cdf[i] = cdf[i-1] + pdf[i]
	}

	return cdf
}
