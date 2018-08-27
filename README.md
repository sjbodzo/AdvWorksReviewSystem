# Adventure Works - Product Reviews

> Purpose: To take in product reviews, approving them according to some rules, notifying the client when the review has been approved or denied. 

### Pre-Requisites
This application is written in Golang natively, using Redis as a job queue and a *modified* form of the Postgres AdventureWorks database for persistence. Users can refer to the included Makefile to invoke the application and its tests. 

Ensure the following:
- You can interact with Makefiles
- You have the Go toolchain installed locally (for running tests)
- You have installed Docker and Docker compose

### Modifications To AdventureWorks

- The included install.sql had broken paths to the csv files copied into the database at launch. This meant the data was not properly loaded into the database when launched via Docker Compose. Fixing this required running two simple regex replacements to fix the paths: 
    - replace `\'\.(\/\w+\.csv\` with `'/data$1`
    - replace `\'(\w+\.csv)\` with `'/data/$1`
- Entries are inserted into `Production.ProductReview`, but they include the primary key ID for the records. This breaks because the SERIAL sequence storing the ID does not account for these ID values already being filled, causing constraint violations on database inserts. To resolve this, the ID values being passed in as primary key IDs were removed.
- Added in unique constraints to ensure the same reviewer cannot review the same product twice, for example:
    ```sql
    UNIQUE (ProductID, EmailAddress),
    UNIQUE (ReviewerName, EmailAddress),
    ```
- Added in a constraint that validates the email format:
    ```sql 
    CONSTRAINT "CK_EmailAddrValid" CHECK (EmailAddress ~* '^[A-Za-z0-9._%-]+@[A-Za-z0-9.-]+[.][A-Za-z]+$')
    ```
- Added foreign key constraint with the `Product` table:
    ```sql
    ALTER TABLE Production.ProductReview ADD
    CONSTRAINT "PK_ProductReview_IDFKey" FOREIGN KEY (ProductID)
    REFERENCES Production.Product(ProductID);
    ```

### Usage
To run the tests: 
```bash
make test
```

To launch the entire application via Docker Compose:
```bash
make run-app
```

An example call with a product review: 

`INPUT:`
```bash
curl -X POST \
  http://localhost:8081/v1/api/reviews \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
	"name": "The Real Donald",
	"email": "theking@daddysmansion.com",
	"productid": 798,
	"review": "This product is gonna be h'\''uge!",
	"rating": 5
}'
```
`RESPONSE:`
```json
{"success":true,"reviewID":6}
```


Bad input returns an HTTP 400 and a slice of the errors:

`INPUT`: 
```bash
curl -X POST \
  http://localhost:8081/v1/api/reviews \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/json' \
  -d '{
	"name": "The Real Donald",
	"email": "theking@foo",
	"review": "This product is gonna be h'\''uge!",
	"rating": 0
}'
```
`RESPONSE`:
```json
{
    "success": false,
    "errors": [
        "Missing param(s): Product ID, Rating",
        "Invalid email address format"
    ]
}
```

### Roadmap
- Deploy via ECS in AWS using Terraform
- Create integration test wrapper via Docker Compose
- API Security features:
    - require JWT token via Auth0 over HTTPS
    - enable CORS restrictions (i.e. same-origin policy)