import { FormInput } from "@/components/form/form-input";
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
import { Loader2 } from "lucide-react";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

interface Variable {
	type: string;
	default: string | number | string[];
	description: string;
}

interface PluginTemplate {
	id: string;
	title: string;
	description: string;
	file: string;
	variables: Record<string, Variable>;
}

interface AddPluginDialogProps {
	isOpen: boolean;
	onOpenChange: (open: boolean) => void;
	onSave: () => void;
}

const variableSchema = z.union([z.string(), z.number(), z.array(z.string())]);

const formSchema = z.object({
	templateId: z.string().min(1),
	variables: z.record(variableSchema),
});

type FormValues = z.infer<typeof formSchema>;

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
		const initialVars: Record<string, string | number | string[]> = {};
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
			<DialogContent className="sm:max-w-[425px]">
				<DialogHeader>
					<DialogTitle>
						{step === "select" ? "Add Plugin" : "Configure Plugin"}
					</DialogTitle>
					<DialogDescription>
						{step === "select"
							? "Choose a plugin template to add"
							: "Configure the plugin variables"}
					</DialogDescription>
				</DialogHeader>

				{step === "select" ? (
					<div className="grid gap-4 py-4">
						{templates?.map((template) => (
							<Button
								key={template.id}
								variant="outline"
								size="lg"
								className="w-full py-10 justify-start"
								onClick={() => handleTemplateSelect(template)}
							>
								<div className="flex flex-col gap-2">
									<div className="font-medium">{template.title}</div>
									<div className="text-sm text-muted-foreground">
										{template.description}
									</div>
								</div>
							</Button>
						))}
					</div>
				) : (
					selectedTemplate && (
						<Form {...form}>
							<form
								onSubmit={form.handleSubmit(onSubmit)}
								className="space-y-4"
							>
								<div className="grid gap-4 py-4">
									{Object.entries(selectedTemplate.variables).map(
										([key, variable]) => (
											<FormInput
												key={key}
												control={form.control}
												name={`variables.${key}`}
												label={key.replace("BUNDECK_", "")}
												description={variable.description}
												type={variable.type === "number" ? "number" : "text"}
											/>
										),
									)}
								</div>
								<DialogFooter>
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
