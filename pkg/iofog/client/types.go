/*
 *  *******************************************************************************
 *  * Copyright (c) 2019 Edgeworx, Inc.
 *  *
 *  * This program and the accompanying materials are made available under the
 *  * terms of the Eclipse Public License v. 2.0 which is available at
 *  * http://www.eclipse.org/legal/epl-2.0
 *  *
 *  * SPDX-License-Identifier: EPL-2.0
 *  *******************************************************************************
 *
 */

package client

type User struct {
	Name     string `json:"firstName"`
	Surname  string `json:"lastName"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ControllerStatus struct {
	Status        string  `json:"status"`
	UptimeSeconds float64 `json:"uptimeSec"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type CreateAgentRequest struct {
	Name    string `json:"name"`
	FogType int32  `json:"fogType"`
}

type CreateAgentResponse struct {
	UUID string
}

type GetAgentProvisionKeyResponse struct {
	Key             string `json:"key"`
	ExpireTimeMsUTC int64  `json:"expirationTime"`
}

type AgentInfo struct {
	UUID                      string  `json:"uuid" yml:"uuid"`
	Name                      string  `json:"name" yml:"name"`
	Location                  string  `json:"location" yml:"location"`
	Latitude                  float64 `json:"latitude" yml:"latitude"`
	Longitude                 float64 `json:"longitude" yml:"longitude"`
	Description               string  `json:"description" yml:"description"`
	DockerURL                 string  `json:"dockerUrl" yml:"dockerUrl"`
	DiskLimit                 int64   `json:"diskLimit" yml:"diskLimit"`
	DiskDirectory             string  `json:"diskDirectory" yml:"diskDirectory"`
	MemoryLimit               int64   `json:"memoryLimit" yml:"memoryLimit"`
	CPULimit                  int64   `json:"cpuLimit" yml:"cpuLimit"`
	LogLimit                  int64   `json:"logLimit" yml:"logLimit"`
	LogDirectory              string  `json:"logDirectory" yml:"logDirectory"`
	LogFileCount              int64   `json:"logFileCount" yml:"logFileCount"`
	StatusFrequency           float64 `json:"statusFrequency" yml:"statusFrequency"`
	ChangeFrequency           float64 `json:"changeFrequency" yml:"changeFrequency"`
	DeviceScanFrequency       float64 `json:"deviceScanFrequency" yml:"deviceScanFrequency"`
	BluetoothEnabled          bool    `json:"bluetoothEnabled" yml:"bluetoothEnabled"`
	WatchdogEnabled           bool    `json:"watchdogEnabled" yml:"watchdogEnabled"`
	AbstractedHardwareEnabled bool    `json:"abstractedHardwareEnabled" yml:"abstractedHardwareEnabled"`
	FogType                   int64   `json:"fogType" yml:"fogType"`
	CreatedTimeRFC3339        string  `json:"created_at" yml:"created"`
	UpdatedTimeRFC3339        string  `json:"updated_at" yml:"updated"`
	LastActive                int64   `json:"lastActive" yml:"lastActive"`
	DaemonStatus              string  `json:"daemonStatus" yml:"daemonStatus"`
	UptimeMs                  int64   `json:"daemonOperatingDuration" yml:"uptime"`
	MemoryUsage               float64 `json:"memoryUsage" yml:"memoryUsage"`
	DiskUsage                 float64 `json:"diskUsage" yml:"diskUsage"`
	CPUUsage                  float64 `json:"cpuUsage" yml:"cpuUsage"`
	MemoryViolation           string  `json:"memoryViolation" yml:"memoryViolation"`
	DiskViolation             string  `json:"diskViolation" yml:"diskViolation"`
	CPUViolation              string  `json:"cpuViolation" yml:"cpuViolation"`
	MicroserviceStatus        string  `json:"microserviceStatus" yml:"microserviceStatus"`
	RepositoryCount           int64   `json:"repositoryCount" yml:"repositoryCount"`
	RepositoryStatus          string  `json:"repositoryStatus" yml:"repositoryStatus"`
	LastStatusTimeMsUTC       int64   `json:"lastStatusTime" yml:"LastStatusTime"`
	IPAddress                 string  `json:"ipAddress" yml:"ipAddress"`
	IPAddressExternal         string  `json:"ipAddressExternal" yml:"ipAddressExternal"`
	ProcessedMessaged         int64   `json:"processedMessages" yml:"ProcessedMessages"`
	MicroserviceMessageCount  int64   `json:"microserviceMessageCounts" yml:"microserviceMessageCount"`
	MessageSpeed              float64 `json:"messageSpeed" yml:"messageSpeed"`
	LastCommandTimeMsUTC      int64   `json:"lastCommandTime" yml:"lastCommandTime"`
	NetworkInterface          string  `json:"networkInterface" yml:"networkInterface"`
	Version                   string  `json:"version" yml:"version"`
	IsReadyToUpgrade          bool    `json:"isReadyToUpgrade" yml:"isReadyToUpgrade"`
	IsReadyToRollback         bool    `json:"isReadyToRollback" yml:"isReadyToRollback"`
	Tunnel                    string  `json:"tunnel" yml:"tunnel"`
}

type ListAgentsResponse struct {
	Agents []AgentInfo `json:"fogs"`
}

type AgentListFilter struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Condition string `json:"condition"`
}

type ConnectorInfo struct {
	IP      string `json:"publicIp"`
	Name    string `json:"name"`
	Domain  string `json:"domain"`
	DevMode bool   `json:"devMode"`
}

type ConnectorInfoList struct {
	Connectors []ConnectorInfo `json:"connectors"`
}
