package model

import "html/template"

type RETURN struct {
	Error      string `json:"error,omitempty"`
	NextAction string `json:"nextAction,omitempty"`
}

type USER struct {
	UserId    string `json:"userId,omitempty"`
	Password  string `json:"password,omitempty"`
	MachineId string `json:"machineId,omitempty"`
	PassCode  string `jon:"passCode,omitempty"`
}

type LITE_USER struct {
	UserId   string `json:"userId,omitempty"`
	Password string `json:"password,omitempty"`
}

type Item struct {
	Key      string
	Name     string
	Selected string
}

type Items struct {
	SelectedItemName         string
	SelectedItemDescription  string
	SelectedItemValue        string
	SelectedItemPurchaseDate string
	SelectedItemSerialNum    string
	MenuImageFile            template.URL
	Message                  string
	ItemList                 []Item
	Pictures                 ImageContainer
}

type ImageContainer struct {
	CurrentFile  string
	DisableLeft  string
	DisableRight string
	DisplayFile  template.URL
}
