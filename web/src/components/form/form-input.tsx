import type { Control, FieldValues, Path } from "react-hook-form";
import {
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "../ui/form";
import { Input, type InputProps } from "../ui/input";

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
					{description && <FormDescription>{description}</FormDescription>}
					<FormMessage />
				</FormItem>
			)}
		/>
	);
}
