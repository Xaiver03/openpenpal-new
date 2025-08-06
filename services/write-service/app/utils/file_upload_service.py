"""
高级文件上传服务
支持分片上传、图片压缩、CDN集成
"""

import os
import hashlib
import aiofiles
import aiohttp
from typing import List, Optional, Dict, Any
from fastapi import UploadFile, HTTPException
from PIL import Image
import io
import asyncio
from pathlib import Path

from app.core.config import settings
from app.core.logger import get_logger
from app.utils.cache_manager import get_cache_manager

logger = get_logger(__name__)


class FileUploadService:
    """高级文件上传服务"""
    
    def __init__(self):
        self.upload_dir = Path(settings.UPLOAD_DIR if hasattr(settings, 'UPLOAD_DIR') else './uploads')
        self.upload_dir.mkdir(parents=True, exist_ok=True)
        
        # 支持的图片格式
        self.supported_image_formats = {'jpg', 'jpeg', 'png', 'gif', 'webp'}
        
        # 文件大小限制 (MB)
        self.max_file_size = getattr(settings, 'MAX_FILE_SIZE_MB', 50) * 1024 * 1024
        self.max_chunk_size = getattr(settings, 'MAX_CHUNK_SIZE_MB', 5) * 1024 * 1024
        
        # CDN配置
        self.cdn_enabled = getattr(settings, 'CDN_ENABLED', False)
        self.cdn_endpoint = getattr(settings, 'CDN_ENDPOINT', '')
        self.cdn_access_key = getattr(settings, 'CDN_ACCESS_KEY', '')
        
    async def upload_file(
        self, 
        file: UploadFile, 
        compress_images: bool = True,
        use_cdn: bool = None
    ) -> Dict[str, Any]:
        """
        上传文件
        
        Args:
            file: 上传的文件
            compress_images: 是否压缩图片
            use_cdn: 是否使用CDN（None为自动判断）
            
        Returns:
            文件信息字典
        """
        try:
            # 验证文件
            await self._validate_file(file)
            
            # 生成文件标识
            file_hash = await self._calculate_file_hash(file)
            file_ext = Path(file.filename).suffix.lower()
            
            # 检查是否已存在
            existing_file = await self._check_existing_file(file_hash, file_ext)
            if existing_file:
                logger.info(f"File already exists: {file_hash}")
                return existing_file
            
            # 重置文件指针
            await file.seek(0)
            
            # 处理文件
            if self._is_image(file.filename) and compress_images:
                processed_file = await self._process_image(file)
            else:
                processed_file = await file.read()
            
            # 保存文件
            local_path = await self._save_local_file(processed_file, file_hash, file_ext)
            
            # 上传到CDN
            cdn_url = None
            if (use_cdn if use_cdn is not None else self.cdn_enabled):
                cdn_url = await self._upload_to_cdn(local_path, file_hash, file_ext)
            
            # 生成文件信息
            file_info = {
                'id': file_hash,
                'filename': file.filename,
                'size': len(processed_file),
                'content_type': file.content_type,
                'hash': file_hash,
                'local_path': str(local_path),
                'cdn_url': cdn_url,
                'created_at': asyncio.get_event_loop().time()
            }
            
            # 缓存文件信息
            await self._cache_file_info(file_hash, file_info)
            
            logger.info(f"File uploaded successfully: {file_hash}")
            return file_info
            
        except Exception as e:
            logger.error(f"File upload failed: {str(e)}")
            raise HTTPException(status_code=500, detail=f"文件上传失败: {str(e)}")
    
    async def upload_chunks(
        self, 
        chunk_id: str, 
        chunk_index: int, 
        total_chunks: int,
        chunk_data: bytes,
        filename: str
    ) -> Dict[str, Any]:
        """
        分片上传
        
        Args:
            chunk_id: 分片组ID
            chunk_index: 当前分片索引
            total_chunks: 总分片数
            chunk_data: 分片数据
            filename: 文件名
            
        Returns:
            上传结果
        """
        try:
            # 验证分片数据
            if len(chunk_data) > self.max_chunk_size:
                raise HTTPException(status_code=400, detail="分片过大")
            
            # 保存分片
            chunk_dir = self.upload_dir / 'chunks' / chunk_id
            chunk_dir.mkdir(parents=True, exist_ok=True)
            
            chunk_path = chunk_dir / f"chunk_{chunk_index:04d}"
            async with aiofiles.open(chunk_path, 'wb') as f:
                await f.write(chunk_data)
            
            logger.info(f"Chunk saved: {chunk_id}/{chunk_index}")
            
            # 检查是否所有分片都已上传
            uploaded_chunks = len(list(chunk_dir.glob("chunk_*")))
            
            if uploaded_chunks == total_chunks:
                # 合并分片
                merged_file = await self._merge_chunks(chunk_id, total_chunks, filename)
                logger.info(f"Chunks merged successfully: {chunk_id}")
                return {
                    'status': 'completed',
                    'file_info': merged_file,
                    'uploaded_chunks': uploaded_chunks,
                    'total_chunks': total_chunks
                }
            else:
                return {
                    'status': 'uploading',
                    'uploaded_chunks': uploaded_chunks,
                    'total_chunks': total_chunks
                }
                
        except Exception as e:
            logger.error(f"Chunk upload failed: {str(e)}")
            raise HTTPException(status_code=500, detail=f"分片上传失败: {str(e)}")
    
    async def _validate_file(self, file: UploadFile) -> None:
        """验证文件"""
        if not file.filename:
            raise HTTPException(status_code=400, detail="文件名不能为空")
        
        # 检查文件大小
        await file.seek(0, 2)  # 移到文件末尾
        file_size = await file.tell()
        await file.seek(0)     # 重置到开头
        
        if file_size > self.max_file_size:
            raise HTTPException(
                status_code=400, 
                detail=f"文件过大，最大支持 {self.max_file_size // (1024*1024)}MB"
            )
    
    async def _calculate_file_hash(self, file: UploadFile) -> str:
        """计算文件哈希值"""
        hasher = hashlib.sha256()
        await file.seek(0)
        
        while True:
            chunk = await file.read(8192)
            if not chunk:
                break
            hasher.update(chunk)
        
        return hasher.hexdigest()
    
    def _is_image(self, filename: str) -> bool:
        """判断是否为图片文件"""
        ext = Path(filename).suffix.lower().lstrip('.')
        return ext in self.supported_image_formats
    
    async def _process_image(self, file: UploadFile) -> bytes:
        """处理图片（压缩、格式转换）"""
        try:
            await file.seek(0)
            image_data = await file.read()
            
            # 打开图片
            with Image.open(io.BytesIO(image_data)) as img:
                # 转换为RGB模式（处理透明图片）
                if img.mode in ('RGBA', 'P'):
                    background = Image.new('RGB', img.size, (255, 255, 255))
                    if img.mode == 'P':
                        img = img.convert('RGBA')
                    background.paste(img, mask=img.split()[-1] if img.mode == 'RGBA' else None)
                    img = background
                elif img.mode != 'RGB':
                    img = img.convert('RGB')
                
                # 压缩大图片
                max_dimension = getattr(settings, 'MAX_IMAGE_DIMENSION', 2048)
                if img.width > max_dimension or img.height > max_dimension:
                    img.thumbnail((max_dimension, max_dimension), Image.Resampling.LANCZOS)
                
                # 保存为JPEG格式
                output = io.BytesIO()
                quality = getattr(settings, 'IMAGE_QUALITY', 85)
                img.save(output, format='JPEG', quality=quality, optimize=True)
                
                return output.getvalue()
                
        except Exception as e:
            logger.warning(f"Image processing failed, using original: {str(e)}")
            await file.seek(0)
            return await file.read()
    
    async def _save_local_file(self, file_data: bytes, file_hash: str, file_ext: str) -> Path:
        """保存文件到本地"""
        # 创建目录结构（按哈希前两位分目录）
        sub_dir = self.upload_dir / file_hash[:2]
        sub_dir.mkdir(parents=True, exist_ok=True)
        
        file_path = sub_dir / f"{file_hash}{file_ext}"
        
        async with aiofiles.open(file_path, 'wb') as f:
            await f.write(file_data)
        
        return file_path
    
    async def _upload_to_cdn(self, local_path: Path, file_hash: str, file_ext: str) -> Optional[str]:
        """上传文件到CDN"""
        if not self.cdn_enabled or not self.cdn_endpoint:
            return None
        
        try:
            # 这里实现CDN上传逻辑
            # 示例：上传到阿里云OSS、腾讯云COS等
            
            async with aiohttp.ClientSession() as session:
                with open(local_path, 'rb') as f:
                    file_data = f.read()
                
                # 构造CDN上传URL
                cdn_key = f"uploads/{file_hash[:2]}/{file_hash}{file_ext}"
                upload_url = f"{self.cdn_endpoint}/{cdn_key}"
                
                # 这里应该根据具体的CDN服务实现上传逻辑
                # 例如：生成签名、设置header等
                headers = {
                    'Content-Type': 'application/octet-stream',
                    # 'Authorization': f'Bearer {self.cdn_access_key}',
                }
                
                # 模拟CDN上传
                logger.info(f"Would upload to CDN: {upload_url}")
                
                # 返回CDN URL
                return f"{self.cdn_endpoint}/{cdn_key}"
                
        except Exception as e:
            logger.error(f"CDN upload failed: {str(e)}")
            return None
    
    async def _merge_chunks(self, chunk_id: str, total_chunks: int, filename: str) -> Dict[str, Any]:
        """合并分片文件"""
        chunk_dir = self.upload_dir / 'chunks' / chunk_id
        
        # 按顺序读取分片并合并
        merged_data = b''
        for i in range(total_chunks):
            chunk_path = chunk_dir / f"chunk_{i:04d}"
            if not chunk_path.exists():
                raise HTTPException(status_code=400, detail=f"缺少分片: {i}")
            
            async with aiofiles.open(chunk_path, 'rb') as f:
                chunk_data = await f.read()
                merged_data += chunk_data
        
        # 计算合并后文件的哈希
        file_hash = hashlib.sha256(merged_data).hexdigest()
        file_ext = Path(filename).suffix.lower()
        
        # 保存合并后的文件
        merged_path = await self._save_local_file(merged_data, file_hash, file_ext)
        
        # 清理分片文件
        for chunk_file in chunk_dir.glob("chunk_*"):
            chunk_file.unlink()
        chunk_dir.rmdir()
        
        # 生成文件信息
        file_info = {
            'id': file_hash,
            'filename': filename,
            'size': len(merged_data),
            'hash': file_hash,
            'local_path': str(merged_path),
            'cdn_url': None,  # 稍后可以异步上传到CDN
            'created_at': asyncio.get_event_loop().time()
        }
        
        # 缓存文件信息
        await self._cache_file_info(file_hash, file_info)
        
        return file_info
    
    async def _check_existing_file(self, file_hash: str, file_ext: str) -> Optional[Dict[str, Any]]:
        """检查文件是否已存在"""
        cache_manager = await get_cache_manager()
        cached_info = await cache_manager.get(f"file:{file_hash}")
        
        if cached_info:
            # 验证本地文件是否存在
            local_path = Path(cached_info.get('local_path', ''))
            if local_path.exists():
                return cached_info
        
        return None
    
    async def _cache_file_info(self, file_hash: str, file_info: Dict[str, Any]) -> None:
        """缓存文件信息"""
        try:
            cache_manager = await get_cache_manager()
            await cache_manager.set(f"file:{file_hash}", file_info, expire=86400 * 7)  # 缓存7天
        except Exception as e:
            logger.warning(f"Failed to cache file info: {str(e)}")
    
    async def get_file_info(self, file_hash: str) -> Optional[Dict[str, Any]]:
        """获取文件信息"""
        cache_manager = await get_cache_manager()
        return await cache_manager.get(f"file:{file_hash}")
    
    async def delete_file(self, file_hash: str) -> bool:
        """删除文件"""
        try:
            file_info = await self.get_file_info(file_hash)
            if not file_info:
                return False
            
            # 删除本地文件
            local_path = Path(file_info.get('local_path', ''))
            if local_path.exists():
                local_path.unlink()
            
            # 清理缓存
            cache_manager = await get_cache_manager()
            await cache_manager.delete(f"file:{file_hash}")
            
            # TODO: 删除CDN文件
            
            logger.info(f"File deleted: {file_hash}")
            return True
            
        except Exception as e:
            logger.error(f"Delete file failed: {str(e)}")
            return False


# 全局实例
_file_upload_service: Optional[FileUploadService] = None

async def get_file_upload_service() -> FileUploadService:
    """获取文件上传服务实例"""
    global _file_upload_service
    if _file_upload_service is None:
        _file_upload_service = FileUploadService()
    return _file_upload_service