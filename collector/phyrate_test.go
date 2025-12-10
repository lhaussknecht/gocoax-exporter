package collector

import (
	"testing"
)

func TestCalculateNPERRate(t *testing.T) {
	tests := []struct {
		name           string
		gapNper        int
		ofdmbNper      int
		fmrPayloadVer  int
		gapVLper       int
		expectedRate   int
	}{
		{
			name:          "MoCA 2.x with VLPER",
			gapNper:       20,
			ofdmbNper:     2000,
			fmrPayloadVer: 0x20,
			gapVLper:      15,
			expectedRate:  (LDPC_LEN_100MHZ * 2000) / ((FFT_LEN_100MHZ + ((20 + 10) * 2)) * 46),
		},
		{
			name:          "MoCA 2.x without VLPER (50MHz)",
			gapNper:       20,
			ofdmbNper:     1000,
			fmrPayloadVer: 0x20,
			gapVLper:      0, // No VLPER, use 50MHz formula
			expectedRate:  (LDPC_LEN_50MHZ * 1000) / ((FFT_LEN_50MHZ + (20*2 + 10)) * 26),
		},
		{
			name:          "Zero gap returns zero",
			gapNper:       0,
			ofdmbNper:     2000,
			fmrPayloadVer: 0x20,
			gapVLper:      15,
			expectedRate:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateNPERRate(tt.gapNper, tt.ofdmbNper, tt.fmrPayloadVer, tt.gapVLper)
			if result != tt.expectedRate {
				t.Errorf("Expected rate %d, got %d", tt.expectedRate, result)
			}
		})
	}
}

func TestCalculateVLPERRate(t *testing.T) {
	tests := []struct {
		name         string
		gapVLper     int
		ofdmbVLper   int
		expectedRate int
	}{
		{
			name:         "Normal VLPER calculation",
			gapVLper:     15,
			ofdmbVLper:   3000,
			expectedRate: (LDPC_LEN_100MHZ * 3000) / ((FFT_LEN_100MHZ + ((15 + 10) * 2)) * 46),
		},
		{
			name:         "Zero gap returns zero",
			gapVLper:     0,
			ofdmbVLper:   3000,
			expectedRate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateVLPERRate(tt.gapVLper, tt.ofdmbVLper)
			if result != tt.expectedRate {
				t.Errorf("Expected rate %d, got %d", tt.expectedRate, result)
			}
		})
	}
}

func TestCalculateGCDRate(t *testing.T) {
	tests := []struct {
		name         string
		gapGcd       int
		ofdmbGcd     int
		mocaNodeVer  int
		expectedRate int
	}{
		{
			name:         "MoCA 2.x GCD (100MHz)",
			gapGcd:       20,
			ofdmbGcd:     2500,
			mocaNodeVer:  0x20,
			expectedRate: (LDPC_LEN_100MHZ * 2500) / ((FFT_LEN_100MHZ + ((20 + 10) * 2)) * 46),
		},
		{
			name:         "MoCA 1.x GCD (50MHz)",
			gapGcd:       15,
			ofdmbGcd:     1200,
			mocaNodeVer:  0x11,
			expectedRate: (LDPC_LEN_50MHZ * 1200) / ((FFT_LEN_50MHZ + (15*2 + 10)) * 26),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateGCDRate(tt.gapGcd, tt.ofdmbGcd, tt.mocaNodeVer)
			if result != tt.expectedRate {
				t.Errorf("Expected rate %d, got %d", tt.expectedRate, result)
			}
		})
	}
}

func TestNewPHYRateMatrix(t *testing.T) {
	matrix := NewPHYRateMatrix()

	if matrix.NPER == nil {
		t.Error("NPER map should be initialized")
	}

	if matrix.VLPER == nil {
		t.Error("VLPER map should be initialized")
	}

	if matrix.GCD == nil {
		t.Error("GCD map should be initialized")
	}
}

func TestMinFunction(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{7, 7, 7},
		{-5, 3, -5},
	}

	for _, tt := range tests {
		result := min(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("min(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
		}
	}
}

// Test with values from the example HTML file
func TestExamplePHYRates(t *testing.T) {
	// From the example screenshot:
	// Node 0->0: 455 (GCD)
	// Node 0->1: 2983 (NPER)
	// Node 1->0: 3488 (NPER)
	// Node 1->1: 701 (GCD)

	// These are approximate tests since we don't have the exact FMR payload values
	// We're just verifying that the calculations produce reasonable results

	// Test that NPER rates are in reasonable range (hundreds to thousands of Mbps)
	rate := CalculateNPERRate(20, 2000, 0x25, 0)
	if rate < 100 || rate > 10000 {
		t.Errorf("NPER rate %d is outside reasonable range (100-10000 Mbps)", rate)
	}

	// Test that GCD rates are typically lower than NPER rates
	gcdRate := CalculateGCDRate(20, 300, 0x25)
	if gcdRate < 0 || gcdRate > 5000 {
		t.Errorf("GCD rate %d is outside reasonable range (0-5000 Mbps)", gcdRate)
	}
}

func TestFMRPayloadParser(t *testing.T) {
	// Create test FMR data
	fmrData := make([]uint32, 50)
	// Fill with dummy data
	for i := range fmrData {
		fmrData[i] = 0x12345678
	}

	nodeVersions := map[int]int{
		0: 0x25,
		1: 0x25,
	}

	parser := NewFMRPayloadParser(
		fmrData,
		0,    // entryNodeID
		0x25, // entryMocaVer
		0x25, // ncMocaVer
		0x25, // mocaNetVer
		0x03, // nodeBitMask (nodes 0 and 1)
		nodeVersions,
	)

	if parser.readIndex != 10 {
		t.Errorf("Expected initial read index 10, got %d", parser.readIndex)
	}

	if !parser.alignmentFlag {
		t.Error("Expected initial alignment flag to be true")
	}

	if parser.entryNodeID != 0 {
		t.Errorf("Expected entry node ID 0, got %d", parser.entryNodeID)
	}
}
