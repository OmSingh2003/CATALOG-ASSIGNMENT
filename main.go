package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

type Point struct {
	X *big.Int
	Y *big.Int
}

type tempRoot struct {
	Base  string `json:"base"`
	Value string `json:"value"`
}

type tempKeys struct {
	N int `json:"n"`
	K int `json:"k"`
}

func parseInputFile(filePath string) ([]Point, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var rawData map[string]json.RawMessage
	if err := json.Unmarshal(fileBytes, &rawData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw json: %w", err)
	}

	var keysData tempKeys
	if err := json.Unmarshal(rawData["keys"], &keysData); err != nil {
		return nil, fmt.Errorf("failed to parse 'keys' object: %w", err)
	}
	k := keysData.K

	points := make([]Point, 0, k)

	for key, rawValue := range rawData {
		if key == "keys" {
			continue
		}

		if len(points) >= k {
			break
		}

		var root tempRoot
		if err := json.Unmarshal(rawValue, &root); err != nil {
			return nil, fmt.Errorf("failed to parse point '%s': %w", key, err)
		}

		x, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid x value (key): %s", key)
		}

		base, err := strconv.Atoi(root.Base)
		if err != nil {
			return nil, fmt.Errorf("invalid base for x=%d: %s", x, root.Base)
		}

		y := new(big.Int)
		_, success := y.SetString(root.Value, base)
		if !success {
			return nil, fmt.Errorf("failed to decode y value for x=%d", x)
		}

		points = append(points, Point{X: big.NewInt(x), Y: y})
	}

	if len(points) < k {
		return nil, fmt.Errorf("not enough points in file: found %d, need %d", len(points), k)
	}

	return points, nil
}

func findSecretC(points []Point) (*big.Int, error) {
	if len(points) == 0 {
		return nil, errors.New("cannot interpolate with zero points")
	}

	secretC := big.NewInt(0)

	numerator := new(big.Int)
	denominator := new(big.Int)
	termNumerator := new(big.Int)
	termValue := new(big.Int)
	negXi := new(big.Int)
	denTerm := new(big.Int)

	for j, pointJ := range points {
		numerator.SetInt64(1)
		denominator.SetInt64(1)

		for i, pointI := range points {
			if i == j {
				continue
			}
			numerator.Mul(numerator, negXi.Neg(pointI.X))
			denominator.Mul(denominator, denTerm.Sub(pointJ.X, pointI.X))
		}

		if denominator.Cmp(big.NewInt(0)) == 0 {
			return nil, fmt.Errorf("interpolation failed: duplicate x-value detected leading to division by zero")
		}

		termNumerator.Mul(pointJ.Y, numerator)
		termValue.Div(termNumerator, denominator)

		secretC.Add(secretC, termValue)
	}

	return secretC, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path_to_json_file>")
		os.Exit(1)
	}
	filePath := os.Args[1]

	points, err := parseInputFile(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully parsed %d points from %s\n", len(points), filePath)

	secretC, err := findSecretC(points)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Printf("\n The calculated secret (c) is: %s\n", secretC.String())
}
