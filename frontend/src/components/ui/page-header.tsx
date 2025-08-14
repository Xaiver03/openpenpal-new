import * as React from "react"
import { cn } from "@/lib/utils"

export interface PageHeaderProps
  extends React.HTMLAttributes<HTMLDivElement> {
  title?: string
  description?: string
}

const PageHeader = React.forwardRef<HTMLDivElement, PageHeaderProps>(
  ({ className, title, description, children, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "flex flex-col space-y-2 pb-6 border-b",
        className
      )}
      {...props}
    >
      {title && (
        <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
      )}
      {description && (
        <p className="text-muted-foreground">{description}</p>
      )}
      {children}
    </div>
  )
)

PageHeader.displayName = "PageHeader"

export { PageHeader }