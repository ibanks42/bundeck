import { Button } from '@/components/ui/button';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import parse from 'html-react-parser';
import { PlusCircle, Trash2 } from 'lucide-react';
import type { Control, Path } from 'react-hook-form';

// Define a generic type for the form values
interface FormArrayInputProps<
  T extends Record<string, string | number | boolean>,
> {
  control: Control<T>;
  name: Path<T>;
  label: string;
  description?: string;
  type?: 'text' | 'number' | 'boolean';
  placeholder?: string;
}

export function FormArrayInput<
  T extends Record<string, string | number | boolean>,
>({
  control,
  name,
  label,
  description,
  type = 'text',
  placeholder,
}: FormArrayInputProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => {
        // Ensure the value is an array
        const values = Array.isArray(field.value) ? field.value : [];

        // Handlers for adding and removing items
        const handleAddItem = () => {
          const defaultValue =
            type === 'number' ? 0 : type === 'boolean' ? false : '';
          field.onChange([...values, defaultValue]);
        };

        const handleRemoveItem = (index: number) => {
          const newValues = [...values];
          newValues.splice(index, 1);
          field.onChange(newValues);
        };

        const handleItemChange = (index: number, value: string | boolean) => {
          const newValues: (string | number | boolean)[] = [...values];
          // Convert to appropriate type
          if (type === 'number') {
            newValues[index] = value === '' ? 0 : Number(value);
          } else if (type === 'boolean') {
            newValues[index] = value === 'true';
          } else {
            newValues[index] = value;
          }
          field.onChange(newValues);
        };

        return (
          <FormItem className='space-y-1'>
            <FormLabel>{label}</FormLabel>
            <div className='space-y-2'>
              {values.map((value, idx) => (
                <div
                  key={`${name}-${idx.toString()}`}
                  className='flex gap-2 items-center'
                >
                  <FormControl>
                    {type === 'boolean' ? (
                      <select
                        className='flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50'
                        value={String(value)}
                        onChange={(e) => handleItemChange(idx, e.target.value)}
                      >
                        <option value='true'>True</option>
                        <option value='false'>False</option>
                      </select>
                    ) : (
                      <Input
                        type={type}
                        placeholder={placeholder}
                        value={value}
                        onChange={(e) => handleItemChange(idx, e.target.value)}
                      />
                    )}
                  </FormControl>
                  <Button
                    type='button'
                    variant='ghost'
                    size='icon'
                    disabled={values.length === 1}
                    onClick={() => handleRemoveItem(idx)}
                  >
                    <Trash2 className='h-4 w-4' />
                  </Button>
                </div>
              ))}
              <Button
                type='button'
                variant='outline'
                size='sm'
                className='mt-2'
                onClick={handleAddItem}
              >
                <PlusCircle className='h-4 w-4 mr-2' />
                Add
              </Button>
            </div>
            {description && (
              <FormDescription>{parse(description)}</FormDescription>
            )}
            <FormMessage />
          </FormItem>
        );
      }}
    />
  );
}
