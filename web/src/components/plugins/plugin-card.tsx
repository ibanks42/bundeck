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
import { useMutation, useQuery } from "@tanstack/react-query";
import { ImageIcon, PlayIcon } from "lucide-react";

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
	const { attributes, listeners, setNodeRef, transform, transition } =
		useSortable({ id: plugin.id });

	const style = {
		transform: CSS.Transform.toString(transform),
		transition,
	};

	const { mutate: runPlugin } = useMutation({
		mutationFn: (plugin: Plugin) => {
			return fetch(`/api/plugins/${plugin.id}/run`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
			});
		},
		onSuccess: async (data: Response) => {
			const text = await data.text();
			toast({
				title: `Plugin ran: ${plugin.name}`,
				description: text,
			});
		},
		onError: (error) => {
			toast({
				title: "Error running plugin",
				description: error.message,
				variant: "destructive",
			});
		},
	});

	const handleCardClick = () => {
		if (!isEditMode) {
			runPlugin(plugin);
		}
	};

	const { data: image } = useQuery({
		queryKey: ["plugin-image", plugin.id],
		queryFn: async () => {
			const response = await fetch(`/api/plugins/${plugin.id}/image`);
			if (!response.ok) {
				return null;
			}
			return response.blob();
		},
		retry: false,
		enabled: !!plugin.id,
	});

	return (
		<div
			ref={setNodeRef}
			style={style}
			{...(isEditMode ? { ...attributes, ...listeners } : {})}
		>
			<Card
				className={cn(
					"w-full",
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
					<CardTitle className="flex items-center gap-2">
						{plugin.name}
					</CardTitle>
				</CardHeader>
				<CardContent className="flex flex-col">
					<div className="flex items-center justify-center overflow-hidden">
						{image ? (
							<img
								src={URL.createObjectURL(image)}
								alt={plugin.name}
								className="size-32 object-contain"
							/>
						) : (
							<div className="flex w-full h-32 bg-muted rounded-lg items-center justify-center">
								<ImageIcon className="size-12 text-muted-foreground" />
							</div>
						)}
					</div>
				</CardContent>
				{isEditMode && (
					<CardFooter className="flex justify-end gap-2">
						<Button onClick={() => runPlugin(plugin)}>
							<PlayIcon className="w-4 h-4 mr-2" />
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
			</Card>
		</div>
	);
}
