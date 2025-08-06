#!/usr/bin/env python3

import os
from dotenv import load_dotenv

# 加载环境变量
load_dotenv()

from app.main import create_app

app = create_app()

if __name__ == '__main__':
    app.run(
        host=app.config.get('HOST', '0.0.0.0'),
        port=app.config.get('PORT', 8004),
        debug=app.config.get('DEBUG', False)
    )