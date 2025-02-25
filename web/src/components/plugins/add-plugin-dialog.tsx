import { FormArrayInput } from "@/components/form/form-array-input";
import { FormBooleanInput } from "@/components/form/form-boolean-input";
import { FormInput } from "@/components/form/form-input";
import {
	Accordion,
	AccordionContent,
	AccordionItem,
	AccordionTrigger,
} from "@/components/ui/accordion";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Form } from "@/components/ui/form";
import { useToast } from "@/hooks/use-toast";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import parse from "html-react-parser";
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import type { Control } from "react-hook-form";
import { useForm } from "react-hook-form";
import { z } from "zod";

interface Variable {
	type: string;
	default: string | number | string[] | boolean | boolean[] | number[];
	description: string;
	label: string;
}

interface PluginTemplate {
	id: string;
	title: string;
	description: string;
	file: string;
	category: string;
	label: string;
	variables: Record<string, Variable>;
}

interface AddPluginDialogProps {
	isOpen: boolean;
	onOpenChange: (open: boolean) => void;
	onSave: () => void;
}

// Define a more precise variable value type
type VariableValue =
	| string
	| number
	| boolean
	| string[]
	| number[]
	| boolean[];

// Define our form values type using z.infer
type FormValues = z.infer<typeof formSchema>;

const variableSchema = z.union([
	z.string(),
	z.number(),
	z.array(z.string()),
	z.array(z.number()),
	z.boolean(),
	z.array(z.boolean()),
]);

const formSchema = z.object({
	templateId: z.string().min(1),
	variables: z.record(variableSchema),
});

// Helper function to determine the input type based on the variable type
function getInputComponent(
	variable: Variable,
	control: Control<FormValues>,
	name: `variables.${string}`, // Make it a template literal type
	key: string,
) {
	// Use the variable's label if available, otherwise fall back to the key
	const label = variable.label || key.replace("BUNDECK_", "");

	// Check if it's an array type
	if (variable.type.endsWith("[]")) {
		// Extract the base type (remove the "[]" suffix)
		const baseType = variable.type.slice(0, -2);
		return (
			<FormArrayInput
				key={name}
				control={
					control as unknown as Control<
						Record<string, string | number | boolean>
					>
				}
				name={name}
				label={label}
				description={variable.description}
				type={baseType as "text" | "number" | "boolean"}
			/>
		);
	}

	// Handle boolean type
	if (variable.type === "boolean") {
		return (
			<FormBooleanInput
				key={name}
				control={control as unknown as Control<Record<string, boolean>>}
				name={name}
				label={label}
				description={variable.description}
			/>
		);
	}

	// Default to standard input
	return (
		<FormInput
			key={name}
			control={control}
			name={name}
			label={label}
			description={variable.description}
			type={variable.type === "number" ? "number" : "text"}
		/>
	);
}

export function AddPluginDialog({
	isOpen,
	onOpenChange,
	onSave,
}: AddPluginDialogProps) {
	const { toast } = useToast();
	const router = useRouter();
	const [step, setStep] = useState<"select" | "configure">("select");
	const [selectedTemplate, setSelectedTemplate] =
		useState<PluginTemplate | null>(null);

	const form = useForm<FormValues>({
		resolver: zodResolver(formSchema),
		defaultValues: {
			templateId: "",
			variables: {},
		},
	});

	// Query for fetching templates
	const { data: templates, isError } = useQuery({
		queryKey: ["plugin-templates"],
		queryFn: async () => {
			const response = await fetch("/api/plugins/templates");
			if (!response.ok) {
				throw new Error("Failed to load plugin templates");
			}
			return response.json() as Promise<PluginTemplate[]>;
		},
		enabled: isOpen,
	});

	// Mutation for creating plugin
	const { mutate, isPending } = useMutation({
		mutationFn: async (values: FormValues) => {
			const response = await fetch("/api/plugins/templates/create", {
				method: "POST",
				headers: {
					"Content-Type": "application/json",
				},
				body: JSON.stringify(values),
			});

			if (!response.ok) {
				const error = await response.json();
				throw new Error(error.error || "Failed to create plugin");
			}

			return response.json();
		},
		onSuccess: () => {
			toast({
				title: "Success",
				description: "Plugin added successfully",
			});
			onSave();
			onOpenChange(false);
			setStep("select");
			setSelectedTemplate(null);
			form.reset();
			router.invalidate();
		},
		onError: (error) => {
			toast({
				title: "Error",
				description: error.message,
				variant: "destructive",
			});
		},
	});

	const handleTemplateSelect = (template: PluginTemplate) => {
		setSelectedTemplate(template);
		// Initialize form with template values
		const initialVars: Record<string, VariableValue> = {};
		for (const [key, variable] of Object.entries(template.variables)) {
			initialVars[key] = variable.default;
		}
		form.reset({
			templateId: template.id,
			variables: initialVars,
		});
		setStep("configure");
	};

	function onSubmit(values: FormValues) {
		mutate(values);
	}

	// Reset form when dialog closes
	useEffect(() => {
		if (!isOpen) {
			setStep("select");
			setSelectedTemplate(null);
			form.reset();
		}
	}, [isOpen, form]);

	if (isError) {
		return (
			<Dialog open={isOpen} onOpenChange={onOpenChange}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>Error</DialogTitle>
						<DialogDescription>
							Failed to load plugin templates
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button onClick={() => onOpenChange(false)}>Close</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		);
	}

	return (
		<Dialog open={isOpen} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-[425px] max-h-[85vh] flex flex-col">
				<DialogHeader className="flex-shrink-0">
					<DialogTitle>
						{step === "select" ? (
							"Add Plugin"
						) : (
							<div className="flex items-center gap-2">
								Configure Plugin
								{selectedTemplate?.label && (
									<div className="px-2 py-0.5 text-xs rounded-md bg-primary/10 text-primary">
										{selectedTemplate.label}
									</div>
								)}
							</div>
						)}
					</DialogTitle>
					<DialogDescription>
						{step === "select"
							? "Choose a plugin template to add"
							: "Configure the plugin variables"}
					</DialogDescription>
				</DialogHeader>

				{step === "select" ? (
					<div className="py-4 px-2 overflow-y-auto pr-1 min-h-[300px] flex-1">
						{templates && templates.length > 0 ? (
							<Accordion type="single" collapsible>
								{/* Group templates by category using Accordion */}
								{Object.entries(
									templates.reduce<Record<string, PluginTemplate[]>>(
										(acc, template) => {
											const category = template.category || "Uncategorized";
											if (!acc[category]) {
												acc[category] = [];
											}
											acc[category].push(template);
											return acc;
										},
										{},
									),
								).map(([category, categoryTemplates]) => (
									<AccordionItem key={category} value={category}>
										<AccordionTrigger>{category}</AccordionTrigger>
										<AccordionContent>
											<div className="space-y-2">
												{categoryTemplates.map((template) => (
													<Button
														key={template.id}
														variant="outline"
														className="w-full h-auto py-3 justify-start text-left border border-border/50 hover:border-border"
														onClick={() => handleTemplateSelect(template)}
													>
														<div className="flex flex-col items-start w-full">
															<div className="flex justify-between w-full items-center">
																<div className="font-medium">
																	{template.title}
																</div>
																{template.label && (
																	<div className="px-2 py-0.5 text-xs rounded-md bg-primary/10 text-primary">
																		{template.label}
																	</div>
																)}
															</div>
															<div className="text-sm text-muted-foreground mt-1">
																{parse(template.description)}
															</div>
														</div>
													</Button>
												))}
											</div>
										</AccordionContent>
									</AccordionItem>
								))}
							</Accordion>
						) : (
							<div className="text-center py-4">
								<p className="text-muted-foreground">
									No plugin templates available
								</p>
							</div>
						)}
					</div>
				) : (
					selectedTemplate && (
						<Form {...form}>
							<form
								onSubmit={form.handleSubmit(onSubmit)}
								className="flex flex-col flex-1"
							>
								<div className="overflow-y-auto py-4 min-h-[300px] flex-1">
									<div className="grid gap-4 px-2">
										{Object.entries(selectedTemplate.variables).map(
											([key, variable]) =>
												getInputComponent(
													variable,
													form.control,
													`variables.${key}`,
													key.replace("BUNDECK_", ""),
												),
										)}
									</div>
								</div>
								<DialogFooter className="flex-shrink-0 pt-2">
									<Button
										type="button"
										variant="outline"
										onClick={() => {
											setStep("select");
											setSelectedTemplate(null);
											form.reset();
										}}
									>
										Back
									</Button>
									<Button type="submit" disabled={isPending}>
										{isPending && (
											<Loader2 className="mr-2 h-4 w-4 animate-spin" />
										)}
										Add Plugin
									</Button>
								</DialogFooter>
							</form>
						</Form>
					)
				)}
			</DialogContent>
		</Dialog>
	);
}
