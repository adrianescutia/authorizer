import _ from 'lodash';

export const hasAdminSecret = () => {
	return (<any>window)['__authorizer__'].isOnboardingCompleted === true;
};

export const capitalizeFirstLetter = (data: string): string =>
	data.charAt(0).toUpperCase() + data.slice(1);

const fallbackCopyTextToClipboard = (text: string) => {
	const textArea = document.createElement('textarea');

	textArea.value = text;
	textArea.style.top = '0';
	textArea.style.left = '0';
	textArea.style.position = 'fixed';

	document.body.appendChild(textArea);
	textArea.focus();
	textArea.select();

	try {
		const successful = document.execCommand('copy');
		const msg = successful ? 'successful' : 'unsuccessful';
		console.log('Fallback: Copying text command was ' + msg);
	} catch (err) {
		console.error('Fallback: Oops, unable to copy', err);
	}
	document.body.removeChild(textArea);
};

export const copyTextToClipboard = (text: string) => {
	if (!navigator.clipboard) {
		fallbackCopyTextToClipboard(text);
		return;
	}
	navigator.clipboard.writeText(text).then(
		() => {
			console.log('Async: Copying to clipboard was successful!');
		},
		(err) => {
			console.error('Async: Could not copy text: ', err);
		}
	);
};

export const getObjectDiff = (obj1: any, obj2: any) => {
	const diff = Object.keys(obj1).reduce((result, key) => {
		if (!obj2.hasOwnProperty(key)) {
			result.push(key);
		} else if (_.isEqual(obj1[key], obj2[key])) {
			const resultKeyIndex = result.indexOf(key);
			result.splice(resultKeyIndex, 1);
		}
		return result;
	}, Object.keys(obj2));

	return diff;
};
