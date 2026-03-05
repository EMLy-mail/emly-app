// Package main provides email reading functionality for EMLy.
// This file contains methods for reading EML, MSG, and PEC email files.
package main

import (
	"time"

	"emly/backend/utils/mail"
)

// =============================================================================
// Email Reading Methods
// =============================================================================

// ReadEML reads a standard .eml file and returns the parsed email data.
func (a *App) ReadEML(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadEML", start, err) }()
	return internal.ReadEmlFile(filePath)
}

// ReadPEC reads a PEC (Posta Elettronica Certificata) .eml file.
func (a *App) ReadPEC(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadPEC", start, err) }()
	return internal.ReadPecInnerEml(filePath)
}

// ReadMSG reads a Microsoft Outlook .msg file and returns the email data.
func (a *App) ReadMSG(filePath string, useExternalConverter bool) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadMSG", start, err) }()
	return internal.ReadMsgFile(filePath)
}

// ReadMSGOSS reads a .msg file using the open-source parser.
func (a *App) ReadMSGOSS(filePath string) (data *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadMSGOSS", start, err) }()
	return internal.ReadMsgFile(filePath)
}

// DetectEmailFormat inspects the file's binary content to determine its format.
func (a *App) DetectEmailFormat(filePath string) (string, error) {
	start := time.Now()
	format, err := internal.DetectEmailFormat(filePath)
	canonicalLog("DetectEmailFormat", start, err)
	return string(format), err
}

// ReadAuto automatically detects the email file format and delegates to the
// appropriate reader.
func (a *App) ReadAuto(filePath string) (result *internal.EmailData, err error) {
	start := time.Now()
	defer func() { canonicalLog("ReadAuto", start, err) }()

	format, err := internal.DetectEmailFormat(filePath)
	if err != nil {
		return nil, err
	}

	switch format {
	case internal.FormatMSG:
		return internal.ReadMsgFile(filePath)
	default: // FormatEML or FormatUnknown – try PEC first, fall back to plain EML
		data, err := internal.ReadPecInnerEml(filePath)
		if err != nil {
			return internal.ReadEmlFile(filePath)
		}
		return data, nil
	}
}

// ShowOpenFileDialog displays the system file picker dialog filtered for email files.
func (a *App) ShowOpenFileDialog() (string, error) {
	return internal.ShowFileDialog(a.ctx)
}
