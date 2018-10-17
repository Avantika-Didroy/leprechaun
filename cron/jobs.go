package cron

import (
	"io/ioutil"
	"time"

	"github.com/kilgaloon/leprechaun/workers"

	"github.com/kilgaloon/leprechaun/recipe"
)

func (c *Cron) buildJobs() {
	files, err := ioutil.ReadDir(c.GetConfig().GetRecipesPath())
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		fullFilepath := c.GetConfig().GetRecipesPath() + "/" + file.Name()
		recipe, err := recipe.Build(fullFilepath)
		if err != nil {
			c.GetLogs().Error(err.Error())
		}
		// recipes that needs to be pushed to queue
		// needs to be schedule by definition
		if recipe.Definition == "cron" {
			c.Service.AddFunc(recipe.Pattern, func() {
				worker, err := c.CreateWorker(&recipe)
				if err != nil {
					switch err {
					case workers.ErrMaxReached:
						c.GetLogs().Info("%s", err)
						go c.processRecipe(&recipe)
					case workers.ErrStillActive:
						c.GetLogs().Info("Worker with NAME: %s is still active", recipe.Name)
					}
					// move this worker to queue and work on it when next worker space is available
					go c.processRecipe(&recipe)
					c.GetLogs().Info("%s", err)
					return
				}

				worker.Run()
			})
		}
	}
}

// ProcessRecipe takes specific recipe and process it
func (c *Cron) processRecipe(r *recipe.Recipe) {
	worker, err := c.CreateWorker(r)
	if err != nil {
		switch err {
		case workers.ErrMaxReached:
			// move this worker to queue and work on it when next worker space is available
			time.Sleep(time.Duration(c.GetConfig().RetryRecipeAfter) * time.Second)
			c.GetLogs().Info("%s, retrying in %d s ...", err, c.GetConfig().RetryRecipeAfter)
			go c.processRecipe(r)
		case workers.ErrStillActive:
			c.GetLogs().Info("Worker with NAME: %s is still active", r.Name)
		}

		return
	}

	c.GetLogs().Info("%s file is in progress... \n", r.Name)
	// worker takeover steps and works on then
	worker.Run()
	//remove lock on client
}
