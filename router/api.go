package router

import (
	"github.com/gin-gonic/gin"
	api "github.com/zijiren233/stable-diffusion-webui-bot/stable-diffusion-webui-api"
)

type I18N struct {
	Key    string
	Value  string
	Groups string
}

type I18NS []*I18N

func uniqueI18NS(slice I18NS) I18NS {
	uniqueMap := make(map[string]bool)
	result := I18NS{}

	for _, val := range slice {
		if _, ok := uniqueMap[val.Key]; !ok {
			uniqueMap[val.Key] = true
			result = append(result, val)
		}
	}

	return result
}

type Any2Img struct {
	api.DrawConfig
	PrePhoto     string `json:"pre_photo,omitempty"`
	ControlPhoto string `json:"control_photo,omitempty"`
}

func apis(eng *gin.Engine) {
	api := eng.Group("/api")
	api.Use(gin.LoggerWithFormatter(log))

	{
		api.GET("/i18n/:code", i18nYaml)
		api.GET("/i18n-json/:code", i18nJson)
	}

	{
		api.GET("/modes", allModels)
		api.GET("/models", models)
		api.GET("/extra-model-groups", extraModelGroups)
		api.GET("/extra-model-groups/:group", extraModelWithGroups)
		api.GET("/extra-models", extraModel)
		api.GET("/control-preprocess", controlPreprocess)
		api.GET("/control-process", controlProcess)
	}

	{
		api.GET("/search-images", searchImages)
		api.GET("/images/:filename", Images)
		rg := api.Group("/search-user-images")
		rg.Use(auth(bot))
		rg.GET("", searchUserImages)
	}

	{
		api.POST("/test-draw-config", testDrawConfig)
		draw := api.Group("/draw").Use(auth(bot))
		draw.POST("", drawPost)
		draw.GET("", drawGet)

		interruptGroup := api.Group("/interrupt").Use(auth(bot))
		interruptGroup.GET("", interrupt)
	}

	{
		ctrlPhoto := api.Group("/detect-ctrl-photo").Use(auth(bot))
		ctrlPhoto.POST("", detectCtrlPhotoPost)
		ctrlPhoto.GET("", detectCtrlPhotoGet).Use(gin.Logger())
	}

	{
		superResolution := api.Group("/super-resolution").Use(auth(bot))
		superResolution.POST("", superResolutionPost)
		superResolution.GET("", superResolutionGet)
	}
}
