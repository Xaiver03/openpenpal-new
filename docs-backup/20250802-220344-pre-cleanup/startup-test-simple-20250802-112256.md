# OpenPenPal Startup Modes Test Report (Simple)
Date: 2025年 8月 2日 星期六 11时22分56秒 CST

## System Information
- Platform: Darwin
- Node Version: v24.2.0
- Go Version: go version go1.24.5 darwin/arm64
- Python Version: Python 3.9.6
- Java Version: The operation couldn’t be completed. Unable to locate a Java Runtime.

## Test Results


### Mode: simple
Start Time: 11:22:56

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:00

---

### Mode: demo
Start Time: 11:23:03

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:06

---

### Mode: development
Start Time: 11:23:09

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:13

---

### Mode: mock
Start Time: 11:23:16

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:19

---

### Mode: production
Start Time: 11:23:22

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:26

---

### Mode: complete
Start Time: 11:23:29

**Status**: ❌ Failed to start or timed out
**Duration**: 0 seconds

**Error Log (last 20 lines):**
```
./test-startup-modes-simple.sh: line 106: timeout: command not found
```

End Time: 11:23:34

---

## Summary

Test completed at: 2025年 8月 2日 星期六 11时23分37秒 CST

### Key Findings
- Simple modes (simple, demo, development, mock) should start quickly
- Complex modes (production, complete) may take longer and some services may fail
- Admin Service (port 8003) is expected to fail if Java is not installed
- Python-based services (Write, OCR) may fail if Python virtual environments are not set up
