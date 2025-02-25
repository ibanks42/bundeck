import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input, type InputProps } from '@/components/ui/input';
import parse from 'html-react-parser';
import type { Control, FieldValues, Path } from 'react-hook-form';

interface FormInputProps<T extends FieldValues> extends InputProps {
  control: Control<T>;
  name: Path<T>;
  label: string;
  description?: string;
}

export function FormInput<T extends FieldValues>({
  control,
  name,
  label,
  description,
  ...props
}: FormInputProps<T>) {
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem>
          <FormLabel>{label}</FormLabel>
          <FormControl>
            <Input {...field} {...props} />
          </FormControl>
          {description && (
            <FormDescription>{parse(description)}</FormDescription>
          )}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
