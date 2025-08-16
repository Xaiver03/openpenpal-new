#!/usr/bin/env node

/**
 * Circular Dependency Checker
 * 循环依赖检查器
 * 
 * Purpose: Automated script to detect and report circular dependencies
 * 目的: 自动脚本检测和报告循环依赖
 * 
 * Usage: npm run check-deps
 */

const fs = require('fs')
const path = require('path')
const { execSync } = require('child_process')

// ================================
// Configuration
// ================================

const CONFIG = {
  // Directories to scan
  scanDirs: [
    'src/lib',
    'src/contexts',
    'src/stores',
    'src/components',
    'src/hooks'
  ],
  
  // File extensions to check
  extensions: ['.ts', '.tsx', '.js', '.jsx'],
  
  // Patterns to ignore
  ignorePatterns: [
    'node_modules',
    '.next',
    'dist',
    'build',
    '*.test.*',
    '*.spec.*',
    '__tests__',
    '*.stories.*'
  ],
  
  // Maximum dependency depth to check
  maxDepth: 10,
  
  // Output options
  output: {
    console: true,
    json: true,
    markdown: true,
    jsonFile: 'circular-deps-report.json',
    markdownFile: 'CIRCULAR-DEPS-REPORT.md'
  }
}

// ================================
// Dependency Graph Builder
// ================================

class DependencyGraph {
  constructor() {
    this.nodes = new Map()
    this.edges = new Map()
    this.cycles = []
  }

  addNode(filePath) {
    if (!this.nodes.has(filePath)) {
      this.nodes.set(filePath, {
        path: filePath,
        imports: new Set(),
        importedBy: new Set()
      })
      this.edges.set(filePath, new Set())
    }
  }

  addEdge(from, to) {
    this.addNode(from)
    this.addNode(to)
    
    this.edges.get(from).add(to)
    this.nodes.get(from).imports.add(to)
    this.nodes.get(to).importedBy.add(from)
  }

  findCycles() {
    const visited = new Set()
    const recursionStack = new Set()
    const cycles = []

    const dfs = (node, path = []) => {
      if (recursionStack.has(node)) {
        // Found a cycle
        const cycleStart = path.indexOf(node)
        if (cycleStart !== -1) {
          const cycle = path.slice(cycleStart).concat([node])
          cycles.push(cycle)
        }
        return
      }

      if (visited.has(node)) {
        return
      }

      visited.add(node)
      recursionStack.add(node)
      path.push(node)

      const edges = this.edges.get(node) || new Set()
      for (const neighbor of edges) {
        dfs(neighbor, [...path])
      }

      recursionStack.delete(node)
      path.pop()
    }

    for (const node of this.nodes.keys()) {
      if (!visited.has(node)) {
        dfs(node)
      }
    }

    this.cycles = cycles
    return cycles
  }

  getStats() {
    return {
      totalFiles: this.nodes.size,
      totalDependencies: Array.from(this.edges.values()).reduce(
        (total, deps) => total + deps.size, 0
      ),
      cyclesFound: this.cycles.length,
      filesInCycles: new Set(this.cycles.flat()).size
    }
  }
}

// ================================
// File Scanner
// ================================

class FileScanner {
  constructor(config) {
    this.config = config
    this.graph = new DependencyGraph()
  }

  shouldIgnore(filePath) {
    return this.config.ignorePatterns.some(pattern => {
      if (pattern.includes('*')) {
        const regex = new RegExp(pattern.replace(/\*/g, '.*'))
        return regex.test(filePath)
      }
      return filePath.includes(pattern)
    })
  }

  hasValidExtension(filePath) {
    return this.config.extensions.some(ext => filePath.endsWith(ext))
  }

  extractImports(filePath, content) {
    const imports = []
    const importRegex = /(?:import|from)\s+['"`]([^'"`]+)['"`]/g
    const requireRegex = /require\s*\(\s*['"`]([^'"`]+)['"`]\s*\)/g
    
    let match

    // Extract ES6 imports
    while ((match = importRegex.exec(content)) !== null) {
      imports.push(match[1])
    }

    // Extract CommonJS requires
    while ((match = requireRegex.exec(content)) !== null) {
      imports.push(match[1])
    }

    return imports.map(imp => this.resolveImport(filePath, imp)).filter(Boolean)
  }

  resolveImport(fromFile, importPath) {
    // Skip external modules (those without relative paths)
    if (!importPath.startsWith('.')) {
      return null
    }

    const baseDir = path.dirname(fromFile)
    const resolved = path.resolve(baseDir, importPath)

    // Try different extensions
    for (const ext of this.config.extensions) {
      const withExt = resolved + ext
      if (fs.existsSync(withExt)) {
        return path.relative(process.cwd(), withExt)
      }

      // Try index files
      const indexFile = path.join(resolved, 'index' + ext)
      if (fs.existsSync(indexFile)) {
        return path.relative(process.cwd(), indexFile)
      }
    }

    // If directory exists, try to find index files
    if (fs.existsSync(resolved) && fs.statSync(resolved).isDirectory()) {
      for (const ext of this.config.extensions) {
        const indexFile = path.join(resolved, 'index' + ext)
        if (fs.existsSync(indexFile)) {
          return path.relative(process.cwd(), indexFile)
        }
      }
    }

    return null
  }

  scanFile(filePath) {
    try {
      const content = fs.readFileSync(filePath, 'utf-8')
      const imports = this.extractImports(filePath, content)
      
      const relativePath = path.relative(process.cwd(), filePath)
      
      imports.forEach(importPath => {
        if (importPath) {
          this.graph.addEdge(relativePath, importPath)
        }
      })
      
      return {
        file: relativePath,
        imports: imports.length,
        success: true
      }
    } catch (error) {
      return {
        file: filePath,
        error: error.message,
        success: false
      }
    }
  }

  async scanDirectory(dirPath) {
    const results = []
    
    const scanRecursive = (currentPath) => {
      if (!fs.existsSync(currentPath)) {
        return
      }

      const items = fs.readdirSync(currentPath)
      
      for (const item of items) {
        const itemPath = path.join(currentPath, item)
        
        if (this.shouldIgnore(itemPath)) {
          continue
        }

        if (fs.statSync(itemPath).isDirectory()) {
          scanRecursive(itemPath)
        } else if (this.hasValidExtension(itemPath)) {
          const result = this.scanFile(itemPath)
          results.push(result)
        }
      }
    }

    scanRecursive(dirPath)
    return results
  }

  async scan() {
    const allResults = []
    
    console.log('🔍 Scanning for circular dependencies...')
    
    for (const dir of this.config.scanDirs) {
      const dirPath = path.join(process.cwd(), dir)
      console.log(`   Scanning: ${dir}`)
      
      const results = await this.scanDirectory(dirPath)
      allResults.push(...results)
    }

    console.log(`📁 Scanned ${allResults.length} files`)
    
    // Find cycles
    const cycles = this.graph.findCycles()
    const stats = this.graph.getStats()
    
    return {
      results: allResults,
      cycles,
      stats,
      graph: this.graph
    }
  }
}

// ================================
// Report Generator
// ================================

class ReportGenerator {
  constructor(config) {
    this.config = config
  }

  generateConsoleReport(scanResult) {
    const { cycles, stats } = scanResult
    
    console.log('\n' + '='.repeat(60))
    console.log('📊 CIRCULAR DEPENDENCY ANALYSIS REPORT')
    console.log('='.repeat(60))
    
    console.log('\n📈 STATISTICS:')
    console.log(`   Total files scanned: ${stats.totalFiles}`)
    console.log(`   Total dependencies: ${stats.totalDependencies}`)
    console.log(`   Circular dependencies found: ${stats.cyclesFound}`)
    console.log(`   Files involved in cycles: ${stats.filesInCycles}`)
    
    if (cycles.length > 0) {
      console.log('\n🔴 CIRCULAR DEPENDENCIES FOUND:')
      cycles.forEach((cycle, index) => {
        console.log(`\n   Cycle ${index + 1}:`)
        cycle.forEach((file, i) => {
          const arrow = i < cycle.length - 1 ? ' → ' : ' ↺ '
          console.log(`     ${file}${arrow}`)
        })
      })
      
      console.log('\n⚠️  ACTION REQUIRED: Please resolve these circular dependencies')
      console.log('💡 SUGGESTIONS:')
      console.log('   - Use dependency injection')
      console.log('   - Extract shared interfaces/types')
      console.log('   - Use dynamic imports (lazy loading)')
      console.log('   - Refactor to remove tight coupling')
    } else {
      console.log('\n✅ NO CIRCULAR DEPENDENCIES FOUND!')
    }
  }

  generateJsonReport(scanResult) {
    const report = {
      timestamp: new Date().toISOString(),
      config: this.config,
      stats: scanResult.stats,
      cycles: scanResult.cycles.map((cycle, index) => ({
        id: index + 1,
        length: cycle.length,
        files: cycle,
        severity: cycle.length > 5 ? 'high' : cycle.length > 3 ? 'medium' : 'low'
      })),
      recommendations: cycles.length > 0 ? [
        'Implement dependency injection',
        'Extract shared types to separate files',
        'Use dynamic imports for heavy dependencies',
        'Consider architectural refactoring'
      ] : ['Maintain current architecture']
    }
    
    if (this.config.output.json) {
      fs.writeFileSync(this.config.output.jsonFile, JSON.stringify(report, null, 2))
      console.log(`\n📄 JSON report saved: ${this.config.output.jsonFile}`)
    }
    
    return report
  }

  generateMarkdownReport(scanResult) {
    const { cycles, stats } = scanResult
    const timestamp = new Date().toISOString()
    
    let markdown = `# Circular Dependencies Report

Generated: ${timestamp}

## Summary

| Metric | Value |
|--------|-------|
| Total Files | ${stats.totalFiles} |
| Total Dependencies | ${stats.totalDependencies} |
| Cycles Found | ${stats.cyclesFound} |
| Files in Cycles | ${stats.filesInCycles} |

`

    if (cycles.length > 0) {
      markdown += `## 🔴 Circular Dependencies Found

${cycles.length} circular dependencies detected:

`
      cycles.forEach((cycle, index) => {
        markdown += `### Cycle ${index + 1}

\`\`\`
${cycle.join(' → ')}}
\`\`\`

**Severity**: ${cycle.length > 5 ? 'High' : cycle.length > 3 ? 'Medium' : 'Low'}
**Files involved**: ${cycle.length}

`
      })
      
      markdown += `## Recommendations

- [ ] Implement dependency injection pattern
- [ ] Extract shared interfaces/types to separate files
- [ ] Use dynamic imports for non-critical dependencies
- [ ] Consider architectural refactoring
- [ ] Add ESLint rules to prevent future cycles

`
    } else {
      markdown += `## ✅ No Circular Dependencies

Great! Your codebase is free from circular dependencies.

### Preventive Measures

- Continue using current architectural patterns
- Consider adding ESLint rules to maintain this state
- Regular dependency analysis as part of CI/CD

`
    }

    if (this.config.output.markdown) {
      fs.writeFileSync(this.config.output.markdownFile, markdown)
      console.log(`📋 Markdown report saved: ${this.config.output.markdownFile}`)
    }
    
    return markdown
  }
}

// ================================
// Main Execution
// ================================

async function main() {
  try {
    console.log('🚀 Starting circular dependency analysis...')
    
    const scanner = new FileScanner(CONFIG)
    const scanResult = await scanner.scan()
    
    const reportGenerator = new ReportGenerator(CONFIG)
    
    // Generate reports
    if (CONFIG.output.console) {
      reportGenerator.generateConsoleReport(scanResult)
    }
    
    if (CONFIG.output.json) {
      reportGenerator.generateJsonReport(scanResult)
    }
    
    if (CONFIG.output.markdown) {
      reportGenerator.generateMarkdownReport(scanResult)
    }
    
    // Exit with error code if cycles found (for CI/CD)
    if (scanResult.cycles.length > 0) {
      console.log('\n❌ Exiting with error code due to circular dependencies')
      process.exit(1)
    } else {
      console.log('\n✅ Analysis complete - no issues found')
      process.exit(0)
    }
    
  } catch (error) {
    console.error('❌ Analysis failed:', error.message)
    process.exit(1)
  }
}

// ================================
// CLI Integration
// ================================

if (require.main === module) {
  // Parse command line arguments
  const args = process.argv.slice(2)
  
  if (args.includes('--help') || args.includes('-h')) {
    console.log(`
Circular Dependency Checker

Usage: node check-circular-deps.js [options]

Options:
  --help, -h          Show this help message
  --quiet, -q         Suppress console output
  --json-only         Generate only JSON report
  --markdown-only     Generate only Markdown report
  --max-depth <n>     Maximum dependency depth to check (default: ${CONFIG.maxDepth})

Examples:
  node check-circular-deps.js
  node check-circular-deps.js --quiet --json-only
  node check-circular-deps.js --max-depth 5
`)
    process.exit(0)
  }
  
  // Process arguments
  if (args.includes('--quiet') || args.includes('-q')) {
    CONFIG.output.console = false
  }
  
  if (args.includes('--json-only')) {
    CONFIG.output.console = false
    CONFIG.output.markdown = false
  }
  
  if (args.includes('--markdown-only')) {
    CONFIG.output.console = false
    CONFIG.output.json = false
  }
  
  const maxDepthIndex = args.indexOf('--max-depth')
  if (maxDepthIndex !== -1 && args[maxDepthIndex + 1]) {
    CONFIG.maxDepth = parseInt(args[maxDepthIndex + 1], 10) || CONFIG.maxDepth
  }
  
  main()
}

// ================================
// Module Export (for programmatic use)
// ================================

module.exports = {
  DependencyGraph,
  FileScanner,
  ReportGenerator,
  CONFIG
}