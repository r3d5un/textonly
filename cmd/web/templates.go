package main

import "textonly.islandwind.me/internal/models"

type templateData struct {
	BlogPost  *models.BlogPost
	BlogPosts []*models.BlogPost
}
