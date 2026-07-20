import Root from "./select.svelte";
import Trigger from "./select-trigger.svelte";
import Content from "./select-content.svelte";
import Item from "./select-item.svelte";
import Group from "./select-group.svelte";
import { Select as SelectPrimitive } from "bits-ui";

const Value = SelectPrimitive.Value;

export {
	Root,
	Trigger,
	Content,
	Item,
	Group,
	Value,
	//
	Root as Select,
	Trigger as SelectTrigger,
	Content as SelectContent,
	Item as SelectItem,
	Group as SelectGroup,
	Value as SelectValue,
};
