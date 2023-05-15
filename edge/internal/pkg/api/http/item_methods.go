package http

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/constants"
	"devais.it/kronos/internal/pkg/db/models"
	"devais.it/kronos/internal/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type itemMethods struct {
	methods
}

func (m *itemMethods) create(c *gin.Context) {
	var items []models.Item
	err := c.BindJSON(&items)

	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	err = services.BatchCreateItems(items, constants.ModifiedByHTTPAPIName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	ids := make([]string, len(items))
	for i := 0; i < len(items); i++ {
		ids[i] = items[i].ID
	}

	if m.replyCreatedData() {
		items, err = services.GetItemsByIDs(ids)
		if err != nil {
			m.writeServiceError(c, err)
			return
		}

		c.JSON(http.StatusCreated, items)
	} else {
		c.JSON(http.StatusCreated, ids)
	}
}

func (m *itemMethods) getByID(c *gin.Context) {
	id := c.Param("item_id")
	item, err := services.GetItemByID(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

//func (m *itemMethods) getByName(c *gin.Context) {
//	name := c.Param("item_name")
//	item, err := services.GetItemByName(name)
//	if err != nil {
//		m.writeServiceError(c, err)
//		return
//	}
//
//	c.JSON(http.StatusOK, item)
//}

func (m *itemMethods) updateByID(c *gin.Context) {
	var patch map[string]interface{}

	id := c.Param("item_id")

	err := c.BindJSON(&patch)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	patch["id"] = id

	err = services.UpdateItem(patch, constants.ModifiedByHTTPAPIName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	item, err := services.GetItemByID(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

func (m *itemMethods) deleteByID(c *gin.Context) {
	id := c.Param("item_id")
	_, hard := c.GetQuery("hard")

	var err error

	if hard {
		err = services.HardDeleteItemByID(id, constants.ModifiedByHTTPAPIName)
	} else {
		err = services.DeleteItemByID(id, constants.ModifiedByHTTPAPIName)
	}

	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (m *itemMethods) getAll(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	items, err := services.GetAllItems(pagination.Page, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (m *itemMethods) getAllByType(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	itemType := c.Param("item_type")

	items, err := services.GetItemsByType(itemType, pagination.PageSize, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (m *itemMethods) findByName(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}
	name := c.Param("item_name")

	items, err := services.FindItemsByName(name, pagination.Page, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (m *itemMethods) findByType(c *gin.Context) {
	var pagination paginationQuery
	err := c.BindQuery(&pagination)
	if err != nil {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}
	itemType := c.Param("item_type")

	items, err := services.FindItemsByType(itemType, pagination.Page, pagination.PageSize)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, items)
}

func (m *itemMethods) count(c *gin.Context) {
	count, err := services.GetItemsCount()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (m *itemMethods) getMac(c *gin.Context) {
	id := c.Param("item_id")
	mac, err := services.GetItemMac(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	if mac.IsZero() {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, mac.ValueOrZero())
}

func (m *itemMethods) getVersion(c *gin.Context) {
	id := c.Param("item_id")
	version, err := services.GetItemVersion(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, version)
}

func (m *itemMethods) getModifiedBy(c *gin.Context) {
	id := c.Param("item_id")
	modifiedBy, err := services.GetItemModifiedBy(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, modifiedBy)
}

func (m *itemMethods) getCustomer(c *gin.Context) {
	id := c.Param("item_id")
	customer, err := services.GetItemCustomerID(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	if customer.IsZero() {
		c.JSON(http.StatusOK, nil)
		return
	}

	c.JSON(http.StatusOK, customer.ValueOrZero())
}

func (m *itemMethods) getChildren(c *gin.Context) {
	id := c.Param("item_id")
	children, err := services.GetItemChildren(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, children)
}

func (m *itemMethods) getParents(c *gin.Context) {
	id := c.Param("item_id")
	parents, err := services.GetItemParents(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, parents)
}

func (m *itemMethods) getRelations(c *gin.Context) {
	id := c.Param("item_id")
	relations, err := services.GetItemRelations(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, relations)
}

func (m *itemMethods) getAttributes(c *gin.Context) {
	id := c.Param("item_id")
	attributes, err := services.GetItemAttributes(id)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributes)
}

func (m *itemMethods) getAttributeByName(c *gin.Context) {
	itemID := c.Param("item_id")
	attributeName := c.Param("attribute_name")
	attribute, err := services.GetItemAttributeByName(itemID, attributeName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attribute)
}

func (m *itemMethods) getAttributeIDByName(c *gin.Context) {
	itemID := c.Param("item_id")
	attributeName := c.Param("attribute_name")
	attributeID, err := services.GetItemAttributeIDByName(itemID, attributeName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributeID)
}

func (m *itemMethods) getAttributeValueByName(c *gin.Context) {
	itemID := c.Param("item_id")
	attributeName := c.Param("attribute_name")
	attributeValue, err := services.GetItemAttributeValueByName(itemID, attributeName)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributeValue)
}

func (m *itemMethods) getAttributesByType(c *gin.Context) {
	itemID := c.Param("item_id")
	attributeType := c.Param("attribute_type")
	attributes, err := services.GetItemAttributesByType(itemID, attributeType)
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, attributes)
}

func newItemMethods(engine *gin.Engine, conf *config.HTTPConfig, rootPath string) *itemMethods {
	g := engine.Group(rootPath)

	m := &itemMethods{
		methods{router: g, conf: conf},
	}

	g.
		POST("/items", m.create).
		GET("/items/count", m.count).
		GET("/item/:item_id", m.getByID).
		//GET("/item/name/:item_name", m.getByName).
		PUT("/item/:item_id", m.updateByID).
		DELETE("/item/:item_id", m.deleteByID).
		GET("/items", m.getAll).
		GET("/items/type/:item_type", m.getAllByType).
		GET("/items/findByName/:item_name", m.findByName).
		GET("/items/findByType/:item_type", m.findByType).
		GET("/item/:item_id/mac", m.getMac).
		GET("/item/:item_id/version", m.getVersion).
		GET("/item/:item_id/modified_by", m.getModifiedBy).
		GET("/item/:item_id/customer", m.getCustomer).
		GET("/item/:item_id/children", m.getChildren).
		GET("/item/:item_id/parents", m.getParents).
		GET("/item/:item_id/relations", m.getRelations).
		GET("/item/:item_id/attributes", m.getAttributes).
		GET("/item/:item_id/attribute/name/:attribute_name", m.getAttributeByName).
		GET("/item/:item_id/attribute/name/:attribute_name/id", m.getAttributeIDByName).
		GET("/item/:item_id/attribute/name/:attribute_name/value", m.getAttributeValueByName).
		GET("/item/:item_id/attributes/type/:attribute_type", m.getAttributesByType)

	return m
}
