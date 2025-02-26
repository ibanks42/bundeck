export interface Plugin {
  id: number;
  name: string;
  code: string;
  image: string;
  image_type: string;
  order_num: number;
  created_at: string;
  updated_at: string;
  run_continuously: boolean;
  interval_seconds: number;
}
