package article_management

import (
	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"net/http"
	"rabbit_gather/src/neo4j_db"
	"rabbit_gather/util"
)

type SearchArticleRequest struct {
	Position  util.Point2D `json:"position" form:"position"  binding:"required"`
	MinRadius float64      `json:"min_radius" form:"min_radius"`
	MaxRadius float64      `json:"max_radius" form:"max_radius" binding:"required"`
}

// SearchArticleHandler will return articles according to specify conditions.
// Se SearchArticleRequest
func (w *ArticleManagement) SearchArticleHandler(c *gin.Context) {
	var searchArticleRequest SearchArticleRequest
	err := c.ShouldBindQuery(&searchArticleRequest)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"err": "wrong input",
		})
		log.DEBUG.Println("wrong input: ", err.Error())
		return
	}
	log.TempLog().Println()
	//log.TempLog().Println(pretty.Sprint(searchArticleRequest))
	//log.TempLog().Println("searchArticleRequest.Position.X: ", searchArticleRequest.Position.X)
	//log.TempLog().Println("25.040056717110396: ", 25.040056717110396)
	session := neo4j_db.GetDriver().NewSession(neo4j.SessionConfig{})
	defer func(session neo4j.Session) {
		e := session.Close()
		if e != nil {
			log.ERROR.Println("error when close session")
		}
	}(session)

	result, err := session.Run(
		util.GetFileStoredPlainText("sql/search_article_with_radius.cyp"),
		map[string]interface{}{
			"longitude":  searchArticleRequest.Position.X,
			"latitude":   searchArticleRequest.Position.Y,
			"min_radius": searchArticleRequest.MinRadius,
			"max_radius": searchArticleRequest.MaxRadius,
		},
	)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"err": "server error",
		})
		log.ERROR.Println("error when neo4j_db script run:", err.Error())
		return
	}
	type Article struct {
		ID        int64        `json:"id"`
		Title     string       `json:"title"`
		Content   string       `json:"content"`
		Timestamp int64        `json:"timestamp"`
		Position  util.Point2D `json:"position"`
		Distance  float64      `json:"distance"`
	}
	var resultArticles []Article
	for result.Next() {
		record := result.Record()
		//pretty.Println("result: ",record)

		article := Article{
			Position: util.Point2D{},
		}
		art, exist := record.Get("article")
		if exist {
			articleProps := art.(neo4j.Node).Props
			if id, exist := articleProps["id"]; exist {
				article.ID = id.(int64)
			}
		}

		position, exist := record.Get("position")
		if exist {
			positionProps := position.(neo4j.Node).Props
			if point, exist := positionProps["pt"]; exist {
				article.Position.Y = point.(neo4j.Point2D).Y
				article.Position.X = point.(neo4j.Point2D).X
			}
		}

		distance, exist := record.Get("distance")
		if exist {
			article.Distance = distance.(float64)
		}

		// Get the article content from MySQL DB
		stat := dbOperator.StatementFromFile("sql/get_article_by_id.sql") //dbOperator.Statement("select title, content,UNIX_TIMESTAMP(update_time) from article as a left join article_tag t on a.id = t.article_id where id = ? and t.tag_id != 1;\n")
		err = stat.QueryRow(article.ID).Scan(&article.Title, &article.Content, &article.Timestamp)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"err": "server error",
			})
			log.ERROR.Println("error when QueryRow:", err.Error())
			return
		}

		resultArticles = append(resultArticles, article)
	}
	c.JSON(200, gin.H{
		"articles": resultArticles,
	})
}
