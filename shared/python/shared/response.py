"""
Shared response utilities for Python services
Safe to import - doesn't affect existing implementations
"""
from typing import Any, Dict
from fastapi import Response
from fastapi.responses import JSONResponse


class APIResponse:
    """Standardized API response utilities"""
    
    @staticmethod
    def success(data: Any = None, message: str = "Success") -> JSONResponse:
        """Return success response"""
        return JSONResponse(
            status_code=200,
            content={
                "success": True,
                "message": message,
                "data": data
            }
        )
    
    @staticmethod
    def error(message: str, status_code: int = 400) -> JSONResponse:
        """Return error response"""
        return JSONResponse(
            status_code=status_code,
            content={
                "success": False,
                "error": message
            }
        )
    
    @staticmethod
    def created(data: Any = None, message: str = "Resource created") -> JSONResponse:
        """Return created response"""
        return JSONResponse(
            status_code=201,
            content={
                "success": True,
                "message": message,
                "data": data
            }
        )
    
    @staticmethod
    def not_found(message: str = "Resource not found") -> JSONResponse:
        """Return not found response"""
        return JSONResponse(
            status_code=404,
            content={
                "success": False,
                "error": message
            }
        )