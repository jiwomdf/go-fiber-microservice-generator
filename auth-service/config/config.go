package config

import (
	"bytes"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	config := *DefaultConfig
	rt := reflect.TypeOf(&config).Elem()
	rv := reflect.ValueOf(&config).Elem()
	lookupEnv("", rt, rv)
	*DefaultConfig = rv.Interface().(Config)
}

func ReadConfig(configFile string) {
	viper.SetConfigName(configFile)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/")
	viper.AddConfigPath("/usr/local/etc/")
	viper.AddConfigPath(".")
	rt := reflect.TypeOf(DefaultConfig).Elem()
	rv := reflect.ValueOf(DefaultConfig).Elem()
	for i := 0; i < rt.NumField(); i++ {
		tag := strings.Split(rt.Field(i).Tag.Get("yaml"), ",")[0]
		name := rt.Field(i).Name
		viper.SetDefault(tag, rv.FieldByName(name).Interface())
	}
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Info("Use default config")
			cfgYAML, err := yaml.Marshal(DefaultConfig)
			if err != nil {
				slog.Error("Failed to marshal YAML", "err", err)
				return
			}
			err = viper.ReadConfig((bytes.NewBuffer(cfgYAML)))
			if err != nil {
				slog.Error("Failed to read config", "err", err)
				return
			}
		} else {
			slog.Error("Config file not found", "err", err)
		}
	} else {
		slog.Info("Use config file", "configFile", configFile)
	}
}

func lookupEnv(parent string, rt reflect.Type, rv reflect.Value) {
	for i := 0; i < rt.NumField(); i++ {
		structField := rt.Field(i)
		tag := strings.Split(structField.Tag.Get("yaml"), ",")[0]
		if structField.Type.Kind() == reflect.Struct {
			lookupEnv(parent+strings.ToUpper(tag)+"_", structField.Type, rv.Field(i))
		} else {
			env := parent + strings.ToUpper(tag)
			value, exist := os.LookupEnv(env)
			if exist {
				slog.Info("ENV", env, value)
				switch structField.Type.Kind().String() {
				case "string":
					rv.Field(i).SetString(value)
				case "bool":
					val, err := strconv.ParseBool(value)
					if err == nil {
						rv.Field(i).SetBool(val)
					}
				case "int", "int8", "int16", "int32", "int64":
					val, err := strconv.ParseInt(value, 10, 64)
					if err == nil {
						rv.Field(i).SetInt(val)
					}
				case "uint", "uint8", "uint16", "uint32", "uint64":
					val, err := strconv.ParseUint(value, 10, 64)
					if err == nil {
						rv.Field(i).SetUint(val)
					}
				case "float32", "float64":
					val, err := strconv.ParseFloat(value, 64)
					if err == nil {
						rv.Field(i).SetFloat(val)
					}
				case "slice":
					values := strings.Split(strings.ReplaceAll(value, " ", ""), ",")
					slice := reflect.MakeSlice(rt.Field(i).Type, len(values), len(values))
					for idx, val := range values {
						switch rt.Field(i).Type.String() {
						case "[]string":
							slice.Index(idx).Set(reflect.ValueOf(val))
						case "[]bool":
							v, err := strconv.ParseBool(val)
							if err == nil {
								slice.Index(idx).Set(reflect.ValueOf(v))
							}
						case "[]int", "[]int8", "[]int16", "[]int32", "[]int64":
							v, err := strconv.ParseInt(val, 10, 64)
							if err == nil {
								slice.Index(idx).Set(reflect.ValueOf(v))
							}
						case "[]uint", "[]uint8", "[]uint16", "[]uint32", "[]uint64":
							v, err := strconv.ParseUint(val, 10, 64)
							if err == nil {
								slice.Index(idx).Set(reflect.ValueOf(v))
							}
						case "[]float32", "[]float64":
							v, err := strconv.ParseFloat(val, 64)
							if err == nil {
								slice.Index(idx).Set(reflect.ValueOf(v))
							}
						}
					}
					rv.Field(i).Set(slice)
				}
			}
		}
	}
}
