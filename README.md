# Go gin crud starter
This repo could be used as a kind of inherited CRUD api, and it will save a lot of time.

## story

I used to handle crud operations in Nodejs with nest crud package, and it is fully compatible with react admin.
I've never spent any time to make sorting or selecting or filtering functionalities. I was able to create
admin dashboard in no time. When I switched to golang I missed these productive tools so I decided to create it.


## Features
- Full crud features
- sorting, selecting and do complex filters like nested "and" & "or" queries
- add joins and nested joins from uri
- compatible with [ra-admin](https://www.radmin.com/)
- inspired from [@nestjsx/crud](https://github.com/nestjsx/crud)

## packages I am using
- Gin
- Gorm
- Postgres
- Air
- Swagger

Fell free to make PR to add other frameworks or ORMs
## How to use

### clone the repo
`git clone git@github.com:ElegantSoft/go-crud-starter.git $your-project-name`

`cd $your-project-name && rm -R .git`

### Configuration
- Create `.env` file for configuration.
- Add to it `DB_URL` as a connection string which is in used by [Gorm](https://gorm.io/docs/connecting_to_the_database.html)
- By default `PORT` is configured on 8080, but you can specify in the `.env` file a different `PORT` if you like so.
- If you want to have the documentation also updated then follow the next steps:
  - Install `swaggo` with `go install github.com/swaggo/swag/cmd/swag@latest`
  - Run in terminal `swag init` which will update the swagger files in the `docs` directory.

### See examples as posts then duplicate it

First you can see how crud api works by running project and go to `/docs/index.html`

- duplicate posts folder
- change package name
- create your entity in models folder
- replace `model` struct in `repository.go`

### requests params
- [Query params](#query-params)
    - [select](#select)
    - [search](#search)
    - [filter conditions](#filter-conditions)
    - [filter](#filter)
    - [or](#or)
    - [sort](#sort)
    - [join](#join)
    - [limit](#limit)
    - [offset](#offset)
    - [page](#page)
    - [cache](#cache)
- [Frontend usage](#frontend-usage)
    - [Customize](#customize)
    - [Usage](#usage)

## Query params

By default, we support these param names:

`fields` - get selected fields in GET result

`s` - search conditions (`$and`, `$or` with all possible variations)

`filter` - filter GET result by `AND` type of condition

`join` - receive joined relational resources in GET result (with all or selected fields)

`sort` - sort GET result by some `field` in `ASC | DESC` order
 
`limit` - limit the amount of received resources

`page` - receive a portion of limited amount of resources


**_Notice:_** You can easily map your own query params names and chose another string delimiters by applying [global options](https://github.com/nestjsx/crud/wiki/Controllers#global-options).

Here is the description of each of those using default params names:

### select

Selects fields that should be returned in the reponse body.

_Syntax:_

> ?fields=**field1**,**field2**,...

_Example:_

> ?fields=**email**,**name**

### search

Adds a search condition as a JSON string to you request. You can combine `$and`, `$or` and use any [condition](#filter-conditions) you need. Make sure it's being sent encoded or just use [`RequestQueryBuilder`](#frontend-usage)

_Syntax:_

> ?s={"name": "Michael"}

_Some examples:_

- Search by field `name` that can be either `null` OR equals `Superman`

> ?s={"name": {"**\$or**": {"**\$isnull**": true, "**\$eq**": "Superman"}}}

- Search an entity where `isActive` is `true` AND `createdAt` not equal `2008-10-01T17:04:32`

> ?s={"**\$and**": [{"isActive": true}, {"createdAt": {"**$ne**": "2008-10-01T17:04:32"}}]}

...which is the same as:

> ?s={"isActive": true, "createdAt": {"**\$ne**": "2008-10-01T17:04:32"}}

- Search an entity where `isActive` is `false` OR `updatedAt` is not `null`

> ?s={"**\$or**": [{"isActive": false}, {"updatedAt": {"**$notnull**": true}}]}

So the amount of combinations is really huge.

**_Notice:_** if search query param is present, then [filter](#filter) and [or](#or) query params will be ignored.

### filter conditions

- **`$eq`** (`=`, equal)
- **`$ne`** (`!=`, not equal)
- **`$gt`** (`>`, greater than)
- **`$lt`** (`<`, lower that)
- **`$gte`** (`>=`, greater than or equal)
- **`$lte`** (`<=`, lower than or equal)
- **`$cont`** (`LIKE %val%`, contains)

### filter

Adds fields request condition (multiple conditions) to your request.

_Syntax:_

> ?filter=**field**||**\$condition**||**value**

> ?join=**relation**&filter=**relation**.**field**||**\$condition**||**value**

**_Notice:_** Using nested filter shall join relation first.

_Examples:_

> ?filter=**name**||**\$eq**||**batman**

> ?filter=**isVillain**||**\$eq**||**false**&filter=**city**||**\$eq**||**Arkham** (multiple filters are treated as a combination of `AND` type of conditions)

> ?filter=**shots**||**\$in**||**12**,**26** (some conditions accept multiple values separated by commas)

> ?filter=**power**||**\$isnull** (some conditions don't accept value)

### or

Adds `OR` conditions to the request.

_Syntax:_

> ?or=**field**||**\$condition**||**value**

It uses the same [filter conditions](#filter-conditions).

_Rules and examples:_

- If there is only **one** `or` present (without `filter`) then it will be interpreted as simple [filter](#filter):

> ?or=**name**||**\$eq**||**batman**

- If there are **multiple** `or` present (without `filter`) then it will be interpreted as a compination of `OR` conditions, as follows:  
  `WHERE {or} OR {or} OR ...`

> ?or=**name**||**\$eq**||**batman**&or=**name**||**\$eq**||**joker**

- If there are **one** `or` and **one** `filter` then it will be interpreted as `OR` condition, as follows:  
  `WHERE {filter} OR {or}`

> ?filter=**name**||**\$eq**||**batman**&or=**name**||**\$eq**||**joker**

- If present **both** `or` and `filter` in any amount (**one** or **miltiple** each) then both interpreted as a combitation of `AND` conditions and compared with each other by `OR` condition, as follows:  
  `WHERE ({filter} AND {filter} AND ...) OR ({or} AND {or} AND ...)`

> ?filter=**type**||**\$eq**||**hero**&filter=**status**||**\$eq**||**alive**&or=**type**||**\$eq**||**villain**&or=**status**||**\$eq**||**dead**

### sort

Adds sort by field (by multiple fields) and order to query result.

_Syntax:_

> ?sort=**field**,**ASC|DESC**

_Examples:_

> ?sort=**name**,**ASC**

> ?sort=**name**,**ASC**&sort=**id**,**DESC**

### join

Receive joined relational objects in GET result (with all or selected fields). You can join as many relations as allowed in your [CrudOptions](https://github.com/nestjsx/crud/wiki/Controllers#join).

_Syntax:_

> ?join=**relation**

> ?join=**relation**||**field1**,**field2**,...

> ?join=**relation1**||**field11**,**field12**,...&join=**relation1**.**nested**||**field21**,**field22**,...&join=...

_Examples:_

> ?join=**profile**

> ?join=**profile**||**firstName**,**email**

> ?join=**profile**||**firstName**,**email**&join=**notifications**||**content**&join=**tasks**

> ?join=**relation1**&join=**relation1**.**nested**&join=**relation1**.**nested**.**deepnested**

**_Notice:_** primary field/column always persists in relational objects. To use nested relations, the parent level **MUST** be set before the child level like example above.

### limit

Receive `N` amount of entities.

_Syntax:_

> ?limit=**number**

_Example:_

> ?limit=**10**

### offset

Limit the amount of received resources

_Syntax:_

> ?offset=**number**

_Example:_

> ?offset=**10**

### page

Receive a portion of limited amount of resources.

_Syntax:_

> ?page=**number**

_Example:_

> ?page=**2**
