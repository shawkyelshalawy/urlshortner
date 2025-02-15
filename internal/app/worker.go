package app

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func (app *Application) StartAnalyticsWorker(ctx context.Context) {
	collection := app.Mongo.Database(app.Config.MongoDatabase).Collection("access_logs")
	
	const batchSize = 100
	const flushInterval = 5 * time.Second
	
	var batch []mongo.WriteModel
	ticker := time.NewTicker(flushInterval)
	
	defer ticker.Stop()

	for {
		select {
		case event := <-app.AnalyticsChan:
			batch = append(batch, mongo.NewInsertOneModel().SetDocument(event))
			if len(batch) >= batchSize {
				app.flushBatch(collection, batch)
				batch = nil
			}

		case <-ticker.C:
			if len(batch) > 0 {
				app.flushBatch(collection, batch)
				batch = nil
			}

		case <-ctx.Done():
			if len(batch) > 0 {
				app.flushBatch(collection, batch)
			}
			return
		}
	}
}

func (app *Application) flushBatch(collection *mongo.Collection, batch []mongo.WriteModel) {
	_, err := collection.BulkWrite(context.Background(), batch)
	if err != nil {
		app.Logger.Error("Failed to write analytics batch", zap.Error(err))
	}
}