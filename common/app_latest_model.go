package common

type AppLatestModel struct {
	AppVersion  string `json:"app_version"`
	AppBuild    string `json:"app_build"`
	AppName     string `json:"app_name"`
	DownloadURL string `json:"download_url"`
	Target      string `json:"target"`
	Arch        string `json:"arch"`
}
