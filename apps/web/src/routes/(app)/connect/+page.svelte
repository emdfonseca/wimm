<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ChooseBankScreen } from '@wimm/ui';

	let { data } = $props();
</script>

<ChooseBankScreen
	banks={data.banks}
	state={data.state}
	onselect={(bankId: string) => goto(resolve('/(app)/connect/[bank]', { bank: bankId }))}
	onretry={() => location.reload()}
/>
