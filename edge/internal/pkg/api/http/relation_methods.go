package http

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

type relationMethods struct {
	methods
}

func (m *relationMethods) create(c *gin.Context) {
	relation := &models.Relation{}

	err := c.BindJSON(relation)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
	}

	err = services.CreateRelation(relation, constants.ModifiedByHTTPAPIName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	if m.replyCreatedData() {
		relation, err = services.GetRelation(relation.ParentID, relation.ChildID)
		if err != nil {
			m.writeServiceError(c, err)
			return
		}
		c.JSON(http.StatusCreated, relation)
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"parent_id": relation.ParentID,
			"child_id":  relation.ChildID,
		})
	}
}

func (m *relationMethods) getByID(c *gin.Context) {
	var query relationQuery

	err := c.ShouldBindWith(&query, binding.Query)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	relation, err := services.GetRelation(query.ParentID, query.ChildID)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, relation)
}

func (m *relationMethods) deleteByID(c *gin.Context) {
	var query relationQuery

	err := c.ShouldBindWith(&query, binding.Query)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	_, hard := c.GetQuery("hard")

	if hard {
		err = services.HardDeleteRelation(query.ParentID, query.ChildID, constants.ModifiedByHTTPAPIName)
	} else {
		err = services.DeleteRelation(query.ParentID, query.ChildID, constants.ModifiedByHTTPAPIName)
	}

	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"parent_id": query.ParentID, "child_id": query.ChildID})
}

func (m *relationMethods) getAll(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	relations, err := services.GetAllRelations(pagination.Page, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, relations)
}

func (m *relationMethods) count(c *gin.Context) {
	count, err := services.GetRelationsCount()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func newRelationMethods(engine *gin.Engine, conf *config.HTTPConfig, rootPath string) *relationMethods {
	g := engine.Group(rootPath)

	m := &relationMethods{
		methods{router: g, conf: conf},
	}

	g.
		POST("/relations", m.create).
		GET("/relations", m.getAll).
		GET("/relation", m.getByID).
		DELETE("/relation", m.deleteByID).
		GET("/relations/count", m.count)

	return m
}
