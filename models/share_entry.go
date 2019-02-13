// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// ShareEntry share entry
// swagger:model ShareEntry
type ShareEntry struct {

	// file ID
	FileID int64 `json:"FileID,omitempty"`

	// ID
	ID int64 `json:"ID,omitempty" gorm:"primary_key;auto_increment"`

	// owner ID
	OwnerID int64 `json:"OwnerID,omitempty" gorm:"-"`

	// shared with ID
	SharedWithID int64 `json:"SharedWithID,omitempty" gorm:"-"`
}

// Validate validates this share entry
func (m *ShareEntry) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ShareEntry) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ShareEntry) UnmarshalBinary(b []byte) error {
	var res ShareEntry
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}