import { cpu } from "node-os-utils";

try {
  const initialPercentage = await cpu.usage();
  console.log(`CPU Usage: ${initialPercentage}%`);
} catch (error) {
  console.error("Error getting initial CPU usage:", error);
}
