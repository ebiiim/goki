package model_test

import (
	"testing"

	"github.com/ebiiim/goki/model"
)

func TestGokiSum_0(t *testing.T) {
	g := model.GokiSum()
	if g.S+g.M+g.L != 0 {
		t.Error("err")
	}
}

func TestGokiSum_1(t *testing.T) {
	g := model.GokiSum(
		model.NewGoki(1, 2, 3),
	)
	if g.S+g.M+g.L != 6 {
		t.Error("err")
	}
}

func TestGokiSum_2(t *testing.T) {
	g := model.GokiSum(
		model.NewGoki(1, 2, 3),
		model.NewGoki(10, 20, 30),
	)
	if g.S+g.M+g.L != 66 {
		t.Error("err")
	}
}

func TestGokiSum_3(t *testing.T) {
	g := model.GokiSum(
		model.NewGoki(1, 2, 3),
		model.NewGoki(10, 20, 30),
		model.NewGoki(100, 200, 300),
	)
	if g.S+g.M+g.L != 666 {
		t.Error("err")
	}
}
