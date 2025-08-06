"""
Minimal FastAPI app for testing basic functionality
"""
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.core.config import settings

# Create FastAPI app
app = FastAPI(
    title="OpenPenPal Write Service",
    description="Write Service for OpenPenPal Platform",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # In production, replace with actual frontend URLs
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/")
async def root():
    """Root endpoint"""
    return {"message": "OpenPenPal Write Service", "version": "1.0.0"}

@app.get("/health")
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": "write-service",
        "timestamp": "2025-07-22T08:00:00Z"
    }

# Add a simple test endpoint
@app.get("/api/test")
async def test_endpoint():
    """Test endpoint"""
    return {"message": "Test endpoint working", "service": "write-service"}

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)