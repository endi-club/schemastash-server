package api

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"context"
	"net/http"
	"schemastash/global"
	"schemastash/types"
	"schemastash/utility"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	Router *gin.Engine
)

func Init() {
	Router = gin.Default()

	Router.GET("/schematics", func(c *gin.Context) {
		var schematics []types.Schematic
		cursor, err := global.Mongo.Collection("schematics").Find(context.TODO(), bson.D{{}})
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
			return
		}
		cursor.All(context.Background(), &schematics)

		for i := range schematics {
			schematics[i].Versions = nil
		}
		c.JSON(http.StatusOK, schematics)
	})

	Router.POST("/schematics", func(c *gin.Context) {
		var schematic types.Schematic
		if c.ShouldBindJSON(&schematic) == nil {
			// check if schematic already exists
			var existingSchematic types.Schematic
			global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": schematic.ID}).Decode(&existingSchematic)
			if existingSchematic.ID != "" {
				c.JSON(http.StatusConflict, map[string]string{"error": "Schematic already exists"})
				return
			}

			schematic.CreatedAt = time.Now().Format(time.RFC3339)
			schematic.Versions = make(map[string]types.Version)

			initialVersion := types.Version{
				ID:          "initial",
				SchematicID: schematic.ID,
				Data:        schematic.Data,
				CreatedAt:   schematic.CreatedAt,
			}

			initialVersionID := utility.VersionHash(schematic, initialVersion)
			initialVersion.ID = initialVersionID
			schematic.LatestVersion = initialVersionID
			schematic.Versions[initialVersionID] = initialVersion

			global.Mongo.Collection("schematics").InsertOne(context.TODO(), schematic)
			c.JSON(http.StatusCreated, schematic)
		} else {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad request"})
		}
	})

	// Get schematic metadata
	Router.GET("/schematics/:schematic_id", func(c *gin.Context) {
		var schematic types.Schematic
		global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&schematic)
		schematic.Versions = nil
		if schematic.ID != "" {
			c.JSON(http.StatusOK, schematic)
		} else {
			c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	})

	Router.DELETE("/schematics/:schematic_id", func(c *gin.Context) {
		var schematic types.Schematic
		global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&schematic)
		if schematic.ID != "" {
			global.Mongo.Collection("schematics").DeleteOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")})
			c.JSON(http.StatusOK, map[string]string{"message": "Schematic deleted"})
		} else {
			c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	})

	Router.PUT("/schematics/:schematic_id", func(c *gin.Context) {
		var schematic types.Schematic
		if c.ShouldBindJSON(&schematic) == nil {
			var existingSchematic types.Schematic
			global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&existingSchematic)
			if existingSchematic.ID != "" {
				schematic.CreatedAt = existingSchematic.CreatedAt
				schematic.Versions = existingSchematic.Versions
				schematic.LatestVersion = existingSchematic.LatestVersion
				global.Mongo.Collection("schematics").ReplaceOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}, schematic)
				c.JSON(http.StatusOK, schematic)
			} else {
				c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
			}
		} else {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad request"})
		}
	})

	// Versioning
	Router.GET("/schematics/:schematic_id/versions", func(c *gin.Context) {
		var schematic types.Schematic
		global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&schematic)
		if schematic.ID != "" {
			c.JSON(http.StatusOK, schematic.Versions)
		} else {
			c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	})

	Router.POST("/schematics/:schematic_id/versions", func(c *gin.Context) {
		var version types.Version
		if c.ShouldBindJSON(&version) == nil {
			var schematic types.Schematic
			global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&schematic)
			if schematic.ID != "" {
				version.SchematicID = schematic.ID
				version.CreatedAt = time.Now().Format(time.RFC3339)
				versionID := utility.VersionHash(schematic, version)
				version.ID = versionID
				schematic.LatestVersion = versionID
				schematic.Data = version.Data
				schematic.Versions[versionID] = version
				global.Mongo.Collection("schematics").ReplaceOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}, schematic)
				c.JSON(http.StatusCreated, version)
			} else {
				c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
			}
		} else {
			c.JSON(http.StatusBadRequest, map[string]string{"error": "Bad request"})
		}
	})

	Router.GET("/schematics/:schematic_id/versions/:version_id", func(c *gin.Context) {
		var schematic types.Schematic
		global.Mongo.Collection("schematics").FindOne(context.TODO(), map[string]string{"id": c.Param("schematic_id")}).Decode(&schematic)
		if schematic.ID != "" {
			version, ok := schematic.Versions[c.Param("version_id")]
			if ok {
				c.JSON(http.StatusOK, version)
			} else {
				c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
			}
		} else {
			c.JSON(http.StatusNotFound, map[string]string{"error": "Not found"})
		}
	})

	// You cannot delete a version.

	Router.Run(":8080")
	log.Println("API server running in port 8080")
}
