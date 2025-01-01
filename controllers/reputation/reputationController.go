package controllers

// var repCollection *mongo.Collection = database.OpenCollection(database.Client, "reputation-tiers")

// // CreateReputation
// func CreateReputation() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		var ctx, cancel = context.WithTimeout(context.Background(), 50*time.Second)
// 		defer cancel()
// 		var rep models.Reputation

// 		if err := c.BindJSON(&rep); err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 			return
// 		}

// 		validationErr := validate.Struct(rep)
// 		if validationErr != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
// 			return
// 		}

// 		filter := bson.D{
// 			{Key: "tier", Value: rep.Tier},
// 		}

// 		count, err := repCollection.CountDocuments(ctx, filter)
// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking reputation existence"})
// 			return
// 		}

// 		if count > 0 {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "rep already exists"})
// 			return
// 		}

// 		rep.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
// 		rep.Updated_at = rep.Created_at
// 		rep.ID = primitive.NewObjectID()

// 		resultInsertionNumber, insertErr := repCollection.InsertOne(ctx, rep)
// 		if insertErr != nil {
// 			msg := "repuatation was not created"
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
// 			return
// 		}

// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "Reputation created successfully",
// 			"succes":  true,
// 			"id":      resultInsertionNumber.InsertedID,
// 		})

// 	}
// }

// func GetAllReputations() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
// 		defer cancel()
// 		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
// 		if err != nil || recordPerPage < 1 {
// 			recordPerPage = 10
// 		}

// 		page, err1 := strconv.Atoi(c.Query("page"))
// 		if err1 != nil || page < 1 {
// 			page = 1
// 		}

// 		startIndex := (page - 1) * recordPerPage

// 		groupStage := bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}}, {Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}}, {Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}}}}}
// 		projectStage := bson.D{
// 			{Key: "$project", Value: bson.D{
// 				{Key: "_id", Value: 0},
// 				{Key: "total_count", Value: 1},
// 				{Key: "reputation_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
// 			}}}

// 		result, err := repCollection.Aggregate(ctx, mongo.Pipeline{
// 			groupStage, projectStage})
// 		defer cancel()

// 		if err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing reputation items"})
// 		}
// 		var data []bson.M
// 		if err = result.All(ctx, &data); err != nil {
// 			log.Fatal(err)
// 		}
// 		c.JSON(http.StatusOK, data[0])

// 	}
// }
