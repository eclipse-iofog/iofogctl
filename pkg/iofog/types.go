package iofog

type User struct {
	Name     string `json:"firstName"`
	Surname  string `json:"lastName"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ControllerStatus struct {
	Status      string `json:"status"`
	UptimeMsUTC int64  `json:"timestamp"`
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
	Key         string `json:"key"`
	ExpireMsUTC int64  `json:"expirationTime"`
}

type GetAgentResponse struct {
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
}
