package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// Side-effect of import configures the postgres driver
	_ "github.com/lib/pq"
)

// NullString alias wraps nullable string in db
type NullString struct {
	sql.NullString
}

// NullInt alias wraps nullable int in db
type NullInt struct {
	sql.NullInt64
}

// NullFloat alias wraps nullable float in db
type NullFloat struct {
	sql.NullFloat64
}

// Wrapper is a pointer to the underlying database library instance
type Wrapper struct {
	_db    *sql.DB
	stmnts map[string]*sql.Stmt
}

// ProductRow is the data in a row of the Product table in the database
type ProductRow struct {
	ProductID             int
	Name                  string
	ProductNumber         string
	MakeFlag              bool
	FinishedGoodsFlag     bool
	Color                 NullString
	SafetyStockLevel      int
	ReorderPoint          int
	StandardCost          int
	ListPrice             int
	Size                  string
	SizeUnitMeasureCode   NullString
	WeightUnitMeasureCode NullString
	Weight                NullFloat
	DaysToManufacture     int
	ProductLine           NullString
	Style                 NullString
	ProductSubcategoryID  NullInt
	ProductModelID        NullInt
	SellStartDate         time.Time
	SellEndDate           time.Time
	DiscontinuedDate      time.Time
}

// ProductReviewRow is the data in a row of the ProductReview table in the database
type ProductReviewRow struct {
	ProductReviewID int
	ProductID       int
	ReviewerName    string
	ReviewDate      time.Time
	EmailAddress    string
	Rating          int
	Comments        *string
	ModifiedDate    time.Time
}

// retryConn retries a db connection 'retries' times every 'wait' duration if it fails
func retryConn(retries int, wait time.Duration, db *sql.DB) (err error) {
	for i := 0; i < retries; i++ {
		if err = db.Ping(); err != nil {
			log.Printf("Database connection failure (%d/%d)\n", i+1, retries)
			time.Sleep(wait)
		} else {
			log.Println("Database connected")
			return nil
		}
	}
	return err
}

// prepareStatements prepares each of the sql statements on the db object
func prepareStatements(db *sql.DB) (statements map[string]*sql.Stmt, err error) {
	statements = make(map[string]*sql.Stmt, 2)

	// Checks for existence of product review in the system by fetching the review's id
	getReviewStmnt, err := db.Prepare("SELECT ProductReviewID FROM Production.ProductReview " +
		"WHERE ProductID=$1 AND ReviewerName=$2 AND EmailAddress=$3")
	if err != nil {
		return nil, err
	}
	statements["GetReview"] = getReviewStmnt

	// Adds new product review into the system
	addReviewStmnt, err := db.Prepare("INSERT INTO Production.ProductReview " +
		"(ProductID, ReviewerName, EmailAddress, Rating, Comments) " +
		"VALUES ($1, $2, $3::varchar(50), $4::smallint, $5) " +
		"RETURNING ProductReviewID")
	if err != nil {
		return nil, err
	}
	statements["AddReview"] = addReviewStmnt

	// Updates existing product review in the system
	updateReviewStmnt, err := db.Prepare("UPDATE Production.ProductReview " +
		"SET Rating=$2::smallint, Comments=$3 WHERE ProductReviewID=$1 RETURNING ProductReviewID")
	if err != nil {
		return nil, err
	}
	statements["UpdateReview"] = updateReviewStmnt

	return statements, nil
}

// closeStatements tries to close the connections for the prepared statements, preventing mem leaks
func (w *Wrapper) closeStatements() (err error) {
	if w._db == nil {
		return
	}

	for _, stmnt := range w.stmnts {
		err := stmnt.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// New returns a new connected db using provided parameters
func New(host string, port int, user string, password string, database string) (*Wrapper, error) {
	c := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%d sslmode=disable ",
		user, database, password, host, port)
	db, err := sql.Open("postgres", c)
	if err != nil {
		return nil, err
	}

	// trigger lazy load of db to guarantee availability
	if err := db.Ping(); err != nil {
		if err := retryConn(3, time.Duration(5)*time.Second, db); err != nil {
			return nil, fmt.Errorf("Could not connect db\nError: %v", err)
		}
	} else {
		log.Println("Database connected")
	}

	// prepare statements to improve db performance
	stmnts, err := prepareStatements(db)
	if err != nil {
		return nil, fmt.Errorf("Unable to prepare statements on db\nError: %v", err)
	}

	return &Wrapper{_db: db, stmnts: stmnts}, nil
}

// Close makes sure we don't leave references/connections dangling
func (w *Wrapper) Close() (err error) {
	if w._db == nil {
		return
	}

	if err = w._db.Close(); err != nil {
		return fmt.Errorf("Unable to close database\nError: %v", err)
	} else if err = w.closeStatements(); err != nil {
		return fmt.Errorf("Unable to close prepared statements on database\nError: %v", err)
	}

	return nil
}

// UpsertReview handles insertions and new additions of product reviews to the database
func (w *Wrapper) UpsertReview(productID int, name string, email string, rating int, comments string) (id int, err error) {
	err = w.stmnts["GetReview"].QueryRow(productID, name, email).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return w.AddReview(productID, name, email, rating, comments)
		}
		return -1, err
	}

	return w.UpdateReview(id, rating, comments)
}

// AddReview adds a new product review to the database
func (w *Wrapper) AddReview(productID int, name string, email string, rating int, comments string) (id int, err error) {
	err = w.stmnts["AddReview"].QueryRow(productID, name, email, rating, comments).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("Unable to add review\nErr: %v", err)
	}

	return id, nil
}

// UpdateReview updates an existing product review in the database
func (w *Wrapper) UpdateReview(reviewID int, rating int, comments string) (id int, err error) {
	err = w.stmnts["UpdateReview"].QueryRow(reviewID, rating, comments).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("Unable to update review\nErr: %v", err)
	}
	return id, nil
}
