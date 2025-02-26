import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardFooter,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";
import type { Plugin } from "@/types/plugin";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useMutation } from "@tanstack/react-query";
import { ImageIcon, PauseIcon, PlayIcon } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Badge } from "../ui/badge";

interface PluginCardProps {
	plugin: Plugin;
	onEdit: (plugin: Plugin) => void;
	onDelete: (plugin: Plugin) => void;
	isEditMode: boolean;
}

export function SortablePluginCard({
	plugin,
	onEdit,
	onDelete,
	isEditMode,
}: PluginCardProps) {
	const { toast } = useToast();
	const [result, setResult] = useState<string | null>(null);
	const [isRunning, setIsRunning] = useState(false);
	const intervalRef = useRef<Timer | null>(null);
	const { attributes, listeners, setNodeRef, transform, transition } =
		useSortable({ id: plugin.id });

	const style = {
		transform: CSS.Transform.toString(transform),
		transition,
		height: "100%",
	};

	const { mutate: runPlugin, isPending } = useMutation({
		mutationFn: (plugin: Plugin) => {
			return fetch(`/api/plugins/${plugin.id}/run`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
			});
		},
		onSuccess: async (data: Response) => {
			const text = await data.text();
			setResult(text);
		},
		onError: (error) => {
			toast({
				title: "Error running plugin",
				description: error.message,
				variant: "destructive",
			});
			// Stop continuous run on error
			if (isRunning) {
				stopContinuousRun();
			}
		},
	});

	const startContinuousRun = useCallback(() => {
		// Don't start if already running or not configured for continuous running
		if (isRunning || !plugin.run_continuously || plugin.interval_seconds <= 0) {
			return;
		}

		// Stop any existing interval first
		stopContinuousRun();

		// Run immediately
		runPlugin(plugin);
		setIsRunning(true);

		try {
			// Use global window.setInterval to ensure we get a numeric ID
			intervalRef.current = setInterval(() => {
				runPlugin(plugin);
			}, plugin.interval_seconds * 1000);
		} catch (err) {
			toast({
				title: "Error setting interval",
				description: (err as Error).message,
				variant: "destructive",
			});
		}
	}, [plugin, runPlugin, isRunning, toast]);

	const stopContinuousRun = useCallback(() => {
		try {
			if (intervalRef.current) {
				window.clearInterval(intervalRef.current);
				intervalRef.current = null;
			}
		} catch (err) {}

		setIsRunning(false);
	}, []);

	// Clean up on unmount
	useEffect(() => {
		return () => {
			if (intervalRef.current) {
				window.clearInterval(intervalRef.current);
				intervalRef.current = null;
			}
		};
	}, []);

	// Handle edit mode changes
	useEffect(() => {
		if (isEditMode && isRunning) {
			stopContinuousRun();
		}
	}, [isEditMode, isRunning, stopContinuousRun]);

	const handleCardClick = () => {
		if (!isEditMode) {
			if (plugin.run_continuously) {
				if (isRunning) {
					stopContinuousRun();
				} else {
					startContinuousRun();
				}
			} else {
				runPlugin(plugin);
			}
		}
	};

	return (
		<div
			ref={setNodeRef}
			style={style}
			{...(isEditMode ? { ...attributes, ...listeners } : {})}
		>
			<Card
				className={cn(
					"w-full h-full flex flex-col",
					isEditMode ? "cursor-move" : "cursor-pointer hover:bg-accent",
				)}
				onClick={handleCardClick}
				tabIndex={0}
				onKeyDown={(e) => {
					if (e.key === "Enter" || e.key === " ") {
						handleCardClick();
					}
				}}
			>
				<CardHeader>
					<CardTitle className="flex items-center justify-between gap-2">
						<span>{plugin.name}</span>
						{plugin.run_continuously && !isEditMode && (
							<Badge variant={isRunning ? "default" : "outline"}>
								{isRunning ? "Running" : "Not Running"}
							</Badge>
						)}
					</CardTitle>
				</CardHeader>
				<CardContent className="flex flex-col gap-2 flex-grow">
					<div className="flex items-center justify-center overflow-hidden">
						{plugin.image ? (
							<img
								src={plugin.image}
								alt={plugin.name}
								className="size-32 object-contain"
							/>
						) : (
							<div className="flex size-32 bg-muted rounded-lg items-center justify-center">
								<ImageIcon className="size-12 text-muted-foreground" />
							</div>
						)}
					</div>

					{((isRunning && plugin.run_continuously) ||
						(result && !plugin.run_continuously)) && (
						<div className="mt-2 w-full">
							<div className="bg-muted p-3 rounded-md mt-1 max-h-32 overflow-y-auto text-sm font-mono whitespace-pre-wrap">
								{result}
							</div>
						</div>
					)}
				</CardContent>
				{isEditMode && (
					<CardFooter className="flex justify-end gap-2 mt-auto">
						<Button onClick={() => runPlugin(plugin)} disabled={isPending}>
							{isPending ? (
								<span className="animate-spin mr-2">⟳</span>
							) : (
								<PlayIcon className="w-4 h-4 mr-2" />
							)}
							Run
						</Button>
						<Button variant="outline" onClick={() => onEdit(plugin)}>
							Edit
						</Button>
						<Button
							variant="destructive"
							onClick={(e) => {
								e.stopPropagation();
								onDelete(plugin);
							}}
						>
							Delete
						</Button>
					</CardFooter>
				)}
				{!isEditMode && plugin.run_continuously && isRunning && (
					<CardFooter className="pt-0 justify-center mt-auto">
						<Button
							variant={isRunning ? "default" : "outline"}
							size="sm"
							onClick={(e) => {
								e.stopPropagation();
								stopContinuousRun();
							}}
						>
							<PauseIcon className="w-4 h-4 mr-2" />
							Stop
						</Button>
					</CardFooter>
				)}
			</Card>
		</div>
	);
}
