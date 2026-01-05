package models

import (
	"errors"

	"github.com/google/uuid"
)

type Engine struct {
	EngineID      uuid.UUID `json:"enigne_id"`
	Displacement  int32     `json:"displacement"`
	NoOfCylinders int32     `json:"noOfCylinders`
	CarRange      int32     `json:"carRange"`
}

type EngineRequest struct {
	Displacement  int32 `json:"displacement"`
	NoOfCylinders int32 `json:"noOfCylinders`
	CarRange      int32 `json:"carRange"`
}

func ValidateEngineRequest(EngineReq EngineRequest) error {
	if err := validateDisplacement(EngineReq.Displacement); err != nil {
		return err
	}

	if err := validateNoOfCylinders(EngineReq.NoOfCylinders); err != nil {
		return err
	}

	if err := validateCarRange(EngineReq.CarRange); err != nil {
		return err
	}
	
	return nil
}

func validateDisplacement(displacement int32) error {
	if displacement <= 0 {
		return errors.New("Displacment must be greater than zero")
	}
	return nil
}

func validateNoOfCylinders(noOfCylinder int32) error {
	if noOfCylinder <= 0 {
		return errors.New("noOfCylinder must be greater than zero")
	}
	return nil
}

func validateCarRange(carRange int32) error {
	if carRange <= 0 {
		return errors.New("noOfCylinder must be greater than zero")
	}
	return nil
}
