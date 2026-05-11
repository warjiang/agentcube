import * as React from "react"
import { cn } from "@/lib/utils"

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'secondary' | 'destructive' | 'outline'
}

function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return (
    <div
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium transition-colors",
        {
          'bg-primary text-primary-foreground': variant === 'default',
          'bg-accent text-accent-foreground': variant === 'secondary',
          'bg-red-100 text-red-700': variant === 'destructive',
          'border border-input bg-background': variant === 'outline',
        },
        className
      )}
      {...props}
    />
  )
}

export { Badge }
