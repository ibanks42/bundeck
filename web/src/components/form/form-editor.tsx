import { Editor, type EditorProps } from '@monaco-editor/react';
import type { Control, FieldValues, Path } from 'react-hook-form';
import { useThemeValue } from '../theme-provider';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '../ui/form';

type FormEditorProps<T extends FieldValues> = {
  control: Control<T>;
  name: Path<T>;
  label?: string;
  description?: string;
  containerClassName?: string;
} & EditorProps;

export function FormEditor<T extends FieldValues>({
  control,
  name,
  label,
  description,
  containerClassName,
  ...props
}: FormEditorProps<T>) {
  const theme = useThemeValue();
  return (
    <FormField
      control={control}
      name={name}
      render={({ field }) => (
        <FormItem className={containerClassName}>
          {label && <FormLabel>{label}</FormLabel>}
          <FormControl>
            <Editor
              height='90%'
              defaultLanguage='typescript'
              language='typescript'
              theme={theme === 'dark' ? 'vs-dark' : 'vs-light'}
              value={field.value}
              onChange={(value) => field.onChange(value || '')}
              options={{
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
              }}
              {...props}
            />
          </FormControl>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
