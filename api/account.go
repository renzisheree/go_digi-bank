package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "tutorial.sqlc.dev/app/db/sqlc"
)

type createAccountRequest struct {
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,currency"`
}

func (server *Server) createAccount(c *gin.Context) {
	var req = createAccountRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
	}
	account, err := server.store.CreateAccount(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getAccount(c *gin.Context) {
	var req = getAccountRequest{}
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccounts(c, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, account)
}

type getListAccountRequest struct {
	PageID   int64 `form:"page_id" binding:"required,min=1"`
	PageSize int64 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) getListAccount(c *gin.Context) {
	var req = getListAccountRequest{}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	args := db.ListAccountsParams{
		Limit:  int32(req.PageSize),
		Offset: int32((req.PageID - 1) * req.PageSize),
	}

	accounts, err := server.store.ListAccounts(c, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, accounts)
}

type UpdateAccountRequest struct {
	ID      int64 `json:"id" binding:"required,min=1"`
	Balance int64 `json:"balance" binding:"required"`
}

func (server *Server) updateAccount(c *gin.Context) {
	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.UpdateAccount(c, db.UpdateAccountParams{
		ID:      req.ID,
		Balance: req.Balance,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, account)
}

type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) deleteAccount(c *gin.Context) {
	var req deleteAccountRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	err := server.store.DeleteAccount(c, req.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	c.JSON(http.StatusOK, "Deleted successfully")

}
