import React from 'react';
import { Check } from 'lucide-react';

interface CheckboxProps {
  id?: string;
  checked?: boolean;
  onCheckedChange?: (checked: boolean) => void;
  className?: string;
  children?: React.ReactNode;
  disabled?: boolean;
}

export const Checkbox: React.FC<CheckboxProps> = ({
  id,
  checked = false,
  onCheckedChange,
  className = '',
  children,
  disabled = false,
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onCheckedChange?.(e.target.checked);
  };

  return (
    <div className={`flex items-center ${className}`}>
      <div className="relative">
        <input
          id={id}
          type="checkbox"
          checked={checked}
          onChange={handleChange}
          disabled={disabled}
          className="sr-only"
        />
        <div
          className={`
            w-4 h-4 border-2 rounded flex items-center justify-center cursor-pointer transition-colors
            ${checked 
              ? 'bg-blue-600 border-blue-600' 
              : 'bg-white border-gray-300 hover:border-gray-400'
            }
          `}
          onClick={() => onCheckedChange?.(!checked)}
        >
          {checked && (
            <Check className="w-3 h-3 text-white" />
          )}
        </div>
      </div>
      {children && (
        <div className="ml-2">
          {children}
        </div>
      )}
    </div>
  );
};