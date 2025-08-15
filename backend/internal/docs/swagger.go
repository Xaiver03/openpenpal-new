package docs

import (
	"github.com/swaggo/swag"
)

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "https://openpenpal.org/terms",
        "contact": {
            "name": "OpenPenPal API Support",
            "url": "https://openpenpal.org/support",
            "email": "support@openpenpal.org"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/v1/auth/register": {
            "post": {
                "description": "Register a new user account",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Authentication"],
                "summary": "Register new user",
                "parameters": [{
                    "description": "User registration data",
                    "name": "request",
                    "in": "body",
                    "required": true,
                    "schema": {
                        "$ref": "#/definitions/RegisterRequest"
                    }
                }],
                "responses": {
                    "200": {
                        "description": "Registration successful",
                        "schema": {
                            "$ref": "#/definitions/AuthResponse"
                        }
                    },
                    "400": {
                        "description": "Invalid request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/auth/login": {
            "post": {
                "description": "Authenticate user and return JWT token",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Authentication"],
                "summary": "User login",
                "parameters": [{
                    "description": "Login credentials",
                    "name": "request",
                    "in": "body",
                    "required": true,
                    "schema": {
                        "$ref": "#/definitions/LoginRequest"
                    }
                }],
                "responses": {
                    "200": {
                        "description": "Login successful",
                        "schema": {
                            "$ref": "#/definitions/AuthResponse"
                        }
                    },
                    "401": {
                        "description": "Invalid credentials",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/letters": {
            "get": {
                "description": "Get paginated list of user's letters",
                "produces": ["application/json"],
                "tags": ["Letters"],
                "summary": "Get user letters",
                "security": [{"JWT": []}],
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page number",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "description": "Items per page",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Filter by status",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Letters retrieved successfully",
                        "schema": {
                            "$ref": "#/definitions/LettersResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new draft letter",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Letters"],
                "summary": "Create draft letter",
                "security": [{"JWT": []}],
                "parameters": [{
                    "description": "Letter data",
                    "name": "request",
                    "in": "body",
                    "required": true,
                    "schema": {
                        "$ref": "#/definitions/CreateLetterRequest"
                    }
                }],
                "responses": {
                    "201": {
                        "description": "Letter created successfully",
                        "schema": {
                            "$ref": "#/definitions/Letter"
                        }
                    }
                }
            }
        },
        "/api/v1/letters/{id}/publish": {
            "post": {
                "description": "Publish a draft letter with optional scheduling",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Letters"],
                "summary": "Publish letter",
                "security": [{"JWT": []}],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Letter ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Publish options",
                        "name": "request",
                        "in": "body",
                        "schema": {
                            "$ref": "#/definitions/PublishLetterRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Letter published successfully",
                        "schema": {
                            "$ref": "#/definitions/Letter"
                        }
                    }
                }
            }
        },
        "/api/v1/courier/apply": {
            "post": {
                "description": "Apply to become a courier",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Courier"],
                "summary": "Apply for courier position",
                "security": [{"JWT": []}],
                "parameters": [{
                    "description": "Courier application data",
                    "name": "request",
                    "in": "body",
                    "required": true,
                    "schema": {
                        "$ref": "#/definitions/CourierApplicationRequest"
                    }
                }],
                "responses": {
                    "200": {
                        "description": "Application submitted successfully",
                        "schema": {
                            "$ref": "#/definitions/CourierApplication"
                        }
                    }
                }
            }
        },
        "/api/v1/museum/entries": {
            "get": {
                "description": "Get paginated museum entries",
                "produces": ["application/json"],
                "tags": ["Museum"],
                "summary": "Get museum entries",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Page number",
                        "name": "page",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Search query",
                        "name": "search",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Museum entries retrieved",
                        "schema": {
                            "$ref": "#/definitions/MuseumEntriesResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/ai/inspiration": {
            "post": {
                "description": "Get AI-generated writing inspiration",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["AI"],
                "summary": "Get writing inspiration",
                "parameters": [{
                    "description": "Inspiration request",
                    "name": "request",
                    "in": "body",
                    "schema": {
                        "$ref": "#/definitions/InspirationRequest"
                    }
                }],
                "responses": {
                    "200": {
                        "description": "Inspiration generated",
                        "schema": {
                            "$ref": "#/definitions/InspirationResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/scheduler/tasks": {
            "get": {
                "description": "Get list of scheduled tasks",
                "produces": ["application/json"],
                "tags": ["Scheduler"],
                "summary": "Get scheduled tasks",
                "security": [{"JWT": []}],
                "parameters": [
                    {
                        "type": "string",
                        "description": "Task type filter",
                        "name": "task_type",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "Task status filter",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "Tasks retrieved successfully",
                        "schema": {
                            "$ref": "#/definitions/ScheduledTasksResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new scheduled task",
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "tags": ["Scheduler"],
                "summary": "Create scheduled task",
                "security": [{"JWT": []}],
                "parameters": [{
                    "description": "Task configuration",
                    "name": "request",
                    "in": "body",
                    "required": true,
                    "schema": {
                        "$ref": "#/definitions/CreateTaskRequest"
                    }
                }],
                "responses": {
                    "201": {
                        "description": "Task created successfully",
                        "schema": {
                            "$ref": "#/definitions/ScheduledTask"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "RegisterRequest": {
            "type": "object",
            "required": ["username", "email", "password"],
            "properties": {
                "username": {
                    "type": "string",
                    "minLength": 3,
                    "maxLength": 50
                },
                "email": {
                    "type": "string",
                    "format": "email"
                },
                "password": {
                    "type": "string",
                    "minLength": 8
                },
                "real_name": {
                    "type": "string"
                },
                "student_id": {
                    "type": "string"
                },
                "school_code": {
                    "type": "string"
                }
            }
        },
        "LoginRequest": {
            "type": "object",
            "required": ["username", "password"],
            "properties": {
                "username": {
                    "type": "string"
                },
                "password": {
                    "type": "string"
                }
            }
        },
        "AuthResponse": {
            "type": "object",
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "token": {
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/User"
                },
                "expires_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "User": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "real_name": {
                    "type": "string"
                },
                "role": {
                    "type": "string",
                    "enum": ["user", "courier", "senior_courier", "courier_coordinator", "school_admin", "platform_admin", "super_admin"]
                },
                "school_code": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "Letter": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                },
                "title": {
                    "type": "string"
                },
                "content": {
                    "type": "string"
                },
                "status": {
                    "type": "string",
                    "enum": ["draft", "scheduled", "published", "delivered", "failed"]
                },
                "visibility": {
                    "type": "string",
                    "enum": ["public", "school", "private"]
                },
                "recipient_op_code": {
                    "type": "string"
                },
                "scheduled_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "CreateLetterRequest": {
            "type": "object",
            "required": ["title", "content"],
            "properties": {
                "title": {
                    "type": "string",
                    "maxLength": 200
                },
                "content": {
                    "type": "string"
                },
                "style": {
                    "type": "string",
                    "enum": ["classic", "modern", "elegant", "casual"]
                },
                "visibility": {
                    "type": "string",
                    "enum": ["public", "school", "private"]
                },
                "recipient_op_code": {
                    "type": "string"
                }
            }
        },
        "PublishLetterRequest": {
            "type": "object",
            "properties": {
                "scheduled_at": {
                    "type": "string",
                    "format": "date-time",
                    "description": "Optional: Schedule letter for future publication"
                },
                "visibility": {
                    "type": "string",
                    "enum": ["public", "school", "private"]
                }
            }
        },
        "LettersResponse": {
            "type": "object",
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "data": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Letter"
                    }
                },
                "pagination": {
                    "$ref": "#/definitions/Pagination"
                }
            }
        },
        "CourierApplicationRequest": {
            "type": "object",
            "required": ["zone", "experience_description"],
            "properties": {
                "zone": {
                    "type": "string",
                    "description": "Zone or area code"
                },
                "experience_description": {
                    "type": "string"
                },
                "availability": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                }
            }
        },
        "CourierApplication": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "user_id": {
                    "type": "string"
                },
                "zone": {
                    "type": "string"
                },
                "status": {
                    "type": "string",
                    "enum": ["pending", "approved", "rejected"]
                },
                "created_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "ScheduledTask": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "task_type": {
                    "type": "string"
                },
                "cron_expression": {
                    "type": "string"
                },
                "status": {
                    "type": "string",
                    "enum": ["pending", "running", "completed", "failed"]
                },
                "next_run_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "last_run_at": {
                    "type": "string",
                    "format": "date-time"
                }
            }
        },
        "CreateTaskRequest": {
            "type": "object",
            "required": ["name", "task_type"],
            "properties": {
                "name": {
                    "type": "string"
                },
                "task_type": {
                    "type": "string"
                },
                "cron_expression": {
                    "type": "string"
                },
                "scheduled_at": {
                    "type": "string",
                    "format": "date-time"
                },
                "payload": {
                    "type": "object"
                }
            }
        },
        "InspirationRequest": {
            "type": "object",
            "properties": {
                "topic": {
                    "type": "string"
                },
                "mood": {
                    "type": "string",
                    "enum": ["happy", "melancholy", "excited", "peaceful", "nostalgic"]
                },
                "length": {
                    "type": "string",
                    "enum": ["short", "medium", "long"]
                }
            }
        },
        "InspirationResponse": {
            "type": "object",
            "properties": {
                "success": {
                    "type": "boolean"
                },
                "inspiration": {
                    "type": "string"
                },
                "prompt_suggestions": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "Pagination": {
            "type": "object",
            "properties": {
                "page": {
                    "type": "integer"
                },
                "limit": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                },
                "total_pages": {
                    "type": "integer"
                }
            }
        },
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "success": {
                    "type": "boolean",
                    "example": false
                },
                "code": {
                    "type": "integer"
                },
                "message": {
                    "type": "string"
                },
                "error": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "JWT": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "JWT token. Format: Bearer {token}"
        }
    }
}`

var doc = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{"http", "https"},
	Title:            "OpenPenPal API",
	Description:      "OpenPenPal Campus Letter Platform API - A comprehensive API for digital campus letter management, courier services, AI features, and museum curation.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(doc.InfoInstanceName, doc)
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}{
	Version:     "1.0.0",
	Host:        "localhost:8080",
	BasePath:    "/",
	Schemes:     []string{"http", "https"},
	Title:       "OpenPenPal API",
	Description: "OpenPenPal Campus Letter Platform API - A comprehensive API for digital campus letter management, courier services, AI features, and museum curation.",
}