package handlers

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wms/config"
)

// setupMockContext membuat gin.Context dengan body JSON dan user_id yang sudah di-set,
// seperti yang dilakukan middleware AuthRequired pada request asli.
func setupMockContext(body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", 1)
	c.Set("role", "admin")

	return c, w
}

func TestCreateTransaction_StockIn_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	config.DB = db

	// Lock row produk, stok awal 10
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT current_stock FROM products WHERE id = $1 FOR UPDATE`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"current_stock"}).AddRow(10))

	// Update stok jadi 10 + 5 = 15
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE products SET current_stock = $1 WHERE id = $2`)).
		WithArgs(15, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Insert ke history transaksi
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO stock_transactions`)).
		WithArgs(1, "IN", 5, "restock", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(100))

	mock.ExpectCommit()

	c, w := setupMockContext(`{"product_id":1,"type":"IN","quantity":5,"note":"restock"}`)
	CreateTransaction(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"new_stock":15`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_StockOut_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	config.DB = db

	// Stok awal 20, ambil 5 -> sisa 15
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT current_stock FROM products WHERE id = $1 FOR UPDATE`)).
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"current_stock"}).AddRow(20))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE products SET current_stock = $1 WHERE id = $2`)).
		WithArgs(15, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO stock_transactions`)).
		WithArgs(2, "OUT", 5, "sale", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(101))

	mock.ExpectCommit()

	c, w := setupMockContext(`{"product_id":2,"type":"OUT","quantity":5,"note":"sale"}`)
	CreateTransaction(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), `"new_stock":15`)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_StockOut_InsufficientStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	config.DB = db

	// Stok cuma 3, tapi mau keluar 10 -> harus ditolak, tidak ada UPDATE/INSERT
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT current_stock FROM products WHERE id = $1 FOR UPDATE`)).
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"current_stock"}).AddRow(3))

	mock.ExpectRollback()

	c, w := setupMockContext(`{"product_id":3,"type":"OUT","quantity":10,"note":"sale"}`)
	CreateTransaction(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Stok tidak cukup")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateTransaction_ProductNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	config.DB = db

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT current_stock FROM products WHERE id = $1 FOR UPDATE`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	c, w := setupMockContext(`{"product_id":999,"type":"IN","quantity":1,"note":""}`)
	CreateTransaction(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.NoError(t, mock.ExpectationsWereMet())
}
