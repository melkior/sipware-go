package api

type User struct {
	DisplayName string
	Email string
	Alias string
}

type RegisterConfig struct {
        From string `json:"From"`
        To string `json:"To"`
        Method string `json:"Method"`
        Alias string `json:"Alias"`
        Password string `json:"Password"`
}

type ContactConfig struct {
        File string `json:"file"`
        Value string `json:"value"`
        Expires string `json:"expires"`
}

type CacheConfig struct {
        Contact ContactConfig `json:"contact"`
}

type Config struct {
        Open string `json:"open"`
        Register RegisterConfig `json:"register"`
        Cache CacheConfig `json:"cache"`
}
