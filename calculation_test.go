package main

import (
	"testing"
)

// Calculator provides basic arithmetic operations
type Calculator struct{}

// Multiply performs multiplication of two integers
func (c *Calculator) Multiply(a, b int) int {
	return a * b // Fixed: Now performs correct multiplication operation
}

// Add performs addition of two integers
func (c *Calculator) Add(a, b int) int {
	return a + b
}

// TestCalculatorMultiply tests the multiplication functionality
func TestCalculatorMultiply(t *testing.T) {
	calc := &Calculator{}

	// Test case 1: Basic multiplication
	result := calc.Multiply(3, 4)
	expected := 12
	if result != expected {
		t.Errorf("Multiply(3, 4): expected %d, got %d", expected, result)
	}

	// Test case 2: Multiplication with zero
	result = calc.Multiply(5, 0)
	expected = 0
	if result != expected {
		t.Errorf("Multiply(5, 0): expected %d, got %d", expected, result)
	}

	// Test case 3: Negative numbers
	result = calc.Multiply(-2, 3)
	expected = -6
	if result != expected {
		t.Errorf("Multiply(-2, 3): expected %d, got %d", expected, result)
	}
}

// TestCalculatorAdd tests the addition functionality (should pass)
func TestCalculatorAdd(t *testing.T) {
	calc := &Calculator{}

	result := calc.Add(3, 4)
	expected := 7
	if result != expected {
		t.Errorf("Add(3, 4): expected %d, got %d", expected, result)
	}
}
