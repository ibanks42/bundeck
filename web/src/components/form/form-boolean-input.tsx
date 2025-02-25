import {
	FormControl,
	FormDescription,
	FormField,
	FormItem,
	FormLabel,
	FormMessage,
} from "@/components/ui/form";
import { Switch } from "@/components/ui/switch";
import parse from "html-react-parser";
import type { Control, Path } from "react-hook-form";

interface FormBooleanInputProps<T extends Record<string, boolean>> {
	control: Control<T>;
	name: Path<T>;
	label: string;
	description?: string;
}

export function FormBooleanInput<T extends Record<string, boolean>>({
	control,
	name,
	label,
	description,
}: FormBooleanInputProps<T>) {
	return (
		<FormField
			control={control}
			name={name}
			render={({ field }) => (
				<FormItem className="flex flex-row items-center justify-between rounded-lg border p-3 shadow-sm">
					<div className="space-y-0.5">
						<FormLabel className="text-base">{label}</FormLabel>
						{description && (
							<FormDescription>{parse(description)}</FormDescription>
						)}
					</div>
					<FormControl>
						<Switch checked={field.value} onCheckedChange={field.onChange} />
					</FormControl>
					<FormMessage />
				</FormItem>
			)}
		/>
	);
}
