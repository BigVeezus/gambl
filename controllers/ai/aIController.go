package controllers

// import (
// 	"github.com/go-playground/validator/v10"
// )

// // var validateUser = validator.New()

// func OpenAiEndpoint() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var aImodel models.AIModel

// 		if err := c.BindJSON(&aImodel); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		validationErr := validateUser.Struct(aImodel)
// 		if validationErr != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}

// 		result, err := config.AskOpenAI(aImodel.Prompt)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		c.JSON(http.StatusOK, result)
// 	}
// }
