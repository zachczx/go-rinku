export default {
	useTabs: true,
	singleQuote: true,
	trailingComma: 'all',
	bracketSameLine: true,
	printWidth: 120,
	plugins: ['prettier-plugin-tailwindcss-extra-plus', 'prettier-plugin-tailwindcss'],
	overrides: [
		{
			files: '*.templ',
			options: { parser: 'tailwindcss-extra-plus' },
		},
	],
};
