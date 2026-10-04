package ai

import (
	"context"
	"sync/atomic"

	"plantpal-backend/internal/domain/plant"
)

var mockIdentifications = []plant.IdentificationResult{
	{
		Species:               "Monstera Deliciosa (Swiss Cheese Plant)",
		SuggestedNickname:     "Monty",
		Location:              "Indoor - Living Room",
		Sunlight:              "Bright indirect sunlight",
		WateringFrequencyDays: 7,
		WaterAmountMl:         250,
		Health:                96,
		CareTips:              "Keep near an east-facing window and let top 2 inches dry between watering.",
	},
	{
		Species:               "Sansevieria Trifasciata (Snake Plant)",
		SuggestedNickname:     "Spiky",
		Location:              "Bedroom / Office Desk",
		Sunlight:              "Low to medium indirect light",
		WateringFrequencyDays: 14,
		WaterAmountMl:         150,
		Health:                98,
		CareTips:              "Very hardy. Extremely drought tolerant, avoid overwatering.",
	},
	{
		Species:               "Epipremnum Aureum (Golden Pothos)",
		SuggestedNickname:     "Leafy",
		Location:              "Balcony / Hanging Basket",
		Sunlight:              "Medium indirect light",
		WateringFrequencyDays: 5,
		WaterAmountMl:         200,
		Health:                94,
		CareTips:              "Water when leaves look slightly limp. Fast-growing and air-purifying.",
	},
	{
		Species:               "Spathiphyllum (Peace Lily)",
		SuggestedNickname:     "Lillian",
		Location:              "Indoor - Shaded Corner",
		Sunlight:              "Low to moderate light",
		WateringFrequencyDays: 6,
		WaterAmountMl:         200,
		Health:                92,
		CareTips:              "Will droop when thirsty, bouncing back quickly after a deep drink.",
	},
}

type MockPlantIdentifier struct {
	counter atomic.Uint64
}

func NewMockPlantIdentifier() *MockPlantIdentifier {
	return &MockPlantIdentifier{}
}

func (p *MockPlantIdentifier) Identify(_ context.Context, _ []byte) (plant.IdentificationResult, error) {
	i := p.counter.Add(1) - 1
	return mockIdentifications[int(i)%len(mockIdentifications)], nil
}
