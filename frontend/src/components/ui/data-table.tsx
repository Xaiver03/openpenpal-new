'use client'

import * as React from 'react'
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  useReactTable,
  SortingState,
  ColumnFiltersState,
  VisibilityState,
  RowSelectionState,
} from '@tanstack/react-table'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight, Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export interface PaginationConfig {
  pageSize: number
  pageIndex: number
  total?: number
}

export interface FilterConfig {
  column: string
  value: string | string[]
  operator?: 'equals' | 'contains' | 'startsWith' | 'endsWith'
}

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  loading?: boolean
  pagination?: PaginationConfig
  onPaginationChange?: (pagination: PaginationConfig) => void
  onSortingChange?: (column: string, direction: 'asc' | 'desc' | null) => void
  onFilterChange?: (filters: FilterConfig[]) => void
  onRowSelectionChange?: (selectedRows: TData[]) => void
  searchKey?: string
  showSearch?: boolean
  showPagination?: boolean
  className?: string
}

export function DataTable<TData, TValue>({
  columns,
  data,
  loading = false,
  pagination,
  onPaginationChange,
  onSortingChange,
  onFilterChange,
  onRowSelectionChange,
  searchKey,
  showSearch = true,
  showPagination = true,
  className,
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = React.useState<SortingState>([])
  const [columnFilters, setColumnFilters] = React.useState<ColumnFiltersState>([])
  const [columnVisibility, setColumnVisibility] = React.useState<VisibilityState>({})
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({})
  const [globalFilter, setGlobalFilter] = React.useState('')

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onSortingChange: (updater: any) => {
      setSorting(updater)
      if (typeof updater === 'function') {
        const newSorting = updater(sorting)
        if (newSorting.length > 0 && onSortingChange) {
          const sort = newSorting[0]
          onSortingChange(sort.id, sort.desc ? 'desc' : 'asc')
        }
      }
    },
    onColumnFiltersChange: (updater: any) => {
      setColumnFilters(updater)
      if (typeof updater === 'function' && onFilterChange) {
        const newFilters = updater(columnFilters)
        onFilterChange(
          newFilters.map((filter: any) => ({
            column: filter.id,
            value: filter.value as string | string[],
          }))
        )
      }
    },
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: (updater: any) => {
      setRowSelection(updater)
      if (typeof updater === 'function' && onRowSelectionChange) {
        const newSelection = updater(rowSelection)
        const selectedRows = data.filter((_, index) => newSelection[index])
        onRowSelectionChange(selectedRows)
      }
    },
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
      globalFilter,
    },
    globalFilterFn: 'includesString',
  })

  // Handle external pagination
  React.useEffect(() => {
    if (pagination) {
      table.setPageSize(pagination.pageSize)
      table.setPageIndex(pagination.pageIndex)
    }
  }, [pagination, table])

  const handlePaginationChange = (pageIndex: number, pageSize: number) => {
    if (onPaginationChange) {
      onPaginationChange({
        pageIndex,
        pageSize,
        total: pagination?.total,
      })
    } else {
      table.setPageIndex(pageIndex)
      table.setPageSize(pageSize)
    }
  }

  return (
    <div className={cn('space-y-4', className)}>
      {showSearch && (
        <div className="flex items-center gap-2">
          <Input
            placeholder={`搜索${searchKey ? ` ${searchKey}` : ''}...`}
            value={globalFilter ?? ''}
            onChange={(event) => setGlobalFilter(event.target.value)}
            className="max-w-sm"
          />
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup: any) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header: any) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  <div className="flex items-center justify-center">
                    <Loader2 className="h-6 w-6 animate-spin" />
                    <span className="ml-2">加载中...</span>
                  </div>
                </TableCell>
              </TableRow>
            ) : table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row: any) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell: any) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  暂无数据
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {showPagination && (
        <div className="flex items-center justify-between px-2">
          <div className="flex-1 text-sm text-muted-foreground">
            {table.getFilteredSelectedRowModel().rows.length > 0 && (
              <span>
                已选择 {table.getFilteredSelectedRowModel().rows.length} / {table.getFilteredRowModel().rows.length} 行
              </span>
            )}
          </div>
          <div className="flex items-center space-x-6 lg:space-x-8">
            <div className="flex items-center space-x-2">
              <p className="text-sm font-medium">每页显示</p>
              <select
                value={table.getState().pagination.pageSize}
                onChange={(e) => {
                  const pageSize = Number(e.target.value)
                  handlePaginationChange(0, pageSize)
                }}
                className="h-8 w-[70px] rounded-md border border-input bg-background px-3 py-1 text-sm"
              >
                {[10, 20, 30, 40, 50].map((pageSize) => (
                  <option key={pageSize} value={pageSize}>
                    {pageSize}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex w-[100px] items-center justify-center text-sm font-medium">
              第 {table.getState().pagination.pageIndex + 1} 页，共 {table.getPageCount()} 页
            </div>
            <div className="flex items-center space-x-2">
              <Button
                variant="outline"
                className="hidden h-8 w-8 p-0 lg:flex"
                onClick={() => handlePaginationChange(0, table.getState().pagination.pageSize)}
                disabled={!table.getCanPreviousPage()}
              >
                <span className="sr-only">首页</span>
                <ChevronsLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                className="h-8 w-8 p-0"
                onClick={() =>
                  handlePaginationChange(
                    table.getState().pagination.pageIndex - 1,
                    table.getState().pagination.pageSize
                  )
                }
                disabled={!table.getCanPreviousPage()}
              >
                <span className="sr-only">上一页</span>
                <ChevronLeft className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                className="h-8 w-8 p-0"
                onClick={() =>
                  handlePaginationChange(
                    table.getState().pagination.pageIndex + 1,
                    table.getState().pagination.pageSize
                  )
                }
                disabled={!table.getCanNextPage()}
              >
                <span className="sr-only">下一页</span>
                <ChevronRight className="h-4 w-4" />
              </Button>
              <Button
                variant="outline"
                className="hidden h-8 w-8 p-0 lg:flex"
                onClick={() =>
                  handlePaginationChange(table.getPageCount() - 1, table.getState().pagination.pageSize)
                }
                disabled={!table.getCanNextPage()}
              >
                <span className="sr-only">末页</span>
                <ChevronsRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}