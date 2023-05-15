package http

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type attributeMethods struct {
	methods
}

func (m *attributeMethods) create(c *gin.Context) {
	attribute := &models.Attribute{}

	err := c.BindJSON(attribute)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	err = services.CreateAttribute(attribute, constants.ModifiedByHTTPAPIName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	if m.replyCreatedData() {
		attribute, err = services.GetAttributeByID(attribute.ID)
		if err != nil {
			m.writeServiceError(c, err)
			return
		}

		c.JSON(http.StatusCreated, attribute)
	} else {
		c.JSON(http.StatusCreated, attribute.ID)
	}
}

func (m *attributeMethods) getByID(c *gin.Context) {
	id := c.Param("id")

	attribute, err := services.GetAttributeByID(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attribute)
}

func (m *attributeMethods) updateByID(c *gin.Context) {
	var patch map[string]interface{}

	id := c.Param("id")

	err := c.BindJSON(&patch)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	patch["id"] = id

	err = services.UpdateAttribute(patch, constants.ModifiedByHTTPAPIName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	attribute, err := services.GetAttributeByID(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attribute)
}

func (m *attributeMethods) deleteByID(c *gin.Context) {
	var err error

	id := c.Param("id")
	_, hard := c.GetQuery("hard")

	if hard {
		err = services.HardDeleteAttributeByID(id, constants.ModifiedByHTTPAPIName)
	} else {
		err = services.DeleteAttributeByID(id, constants.ModifiedByHTTPAPIName)
	}

	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (m *attributeMethods) getAll(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	attributes, err := services.GetAllAttributes(pagination.Page, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributes)
}

func (m *attributeMethods) count(c *gin.Context) {
	count, err := services.GetAttributesCount()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (m *attributeMethods) getValue(c *gin.Context) {
	attributeID := c.Param("id")
	value, err := services.GetAttributeValue(attributeID)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, value)
}

func (m *attributeMethods) getByType(c *gin.Context) {
	attributeType := c.Param("attribute_type")

	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	attributes, err := services.GetAttributesByType(
		attributeType,
		pagination.Page,
		pagination.PageSize,
	)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributes)
}

func newAttributeMethods(engine *gin.Engine, conf *config.HTTPConfig, rootPath string) *attributeMethods {
	g := engine.Group(rootPath)

	m := &attributeMethods{
		methods{router: g, conf: conf},
	}

	g.
		POST("/attributes", m.create).
		GET("/attributes", m.getAll).
		GET("/attributes/type/:attribute_type", m.getByType).
		GET("/attribute/:id", m.getByID).
		PUT("/attribute/:id", m.updateByID).
		GET("/attribute/:id/value", m.getValue).
		DELETE("/attribute/:id", m.deleteByID).
		GET("/attributes/count", m.count)

	return m
}
