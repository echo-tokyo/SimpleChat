import { useCallback, useEffect, useRef, useState } from 'react'
import { SubmitHandler, useForm } from 'react-hook-form'
import { useChatStore } from '../../../store/store'

const Footer = () => {
	const nickname = localStorage.getItem('registered') as string
	const addMessage = useChatStore(state => state.addMessage)
	const [value, setValue] = useState('')
	const [submitState, setSubmitState] = useState(true)
	const textareaElement = useRef<HTMLTextAreaElement>(null)
	const formElement = useRef<HTMLFormElement>(null)
	const { handleSubmit, register, setValue: setFormValue, reset } = useForm<{ content: string }>({
		defaultValues: {
			content: ''
		}
	})

	// Очистка поля, получение данных
	const onSubmit: SubmitHandler<{ content: string }> = useCallback((data) => {
		addMessage( { content: data.content, sender: {username: nickname}, createdAt: new Date().toISOString() } )
		reset()
		formElement.current?.reset()
		setSubmitState(!submitState)
		setValue('')
	}, [addMessage, nickname, reset, submitState])

	// Сохранение значений поля
	const handleChange = (e: React.FormEvent<HTMLTextAreaElement>) => {
		const newValue = (e.target as HTMLTextAreaElement).value
		setValue(newValue)
		setFormValue('content', newValue)
	}

	// Ctrl + shift - перенос строки, enter - отправка
	const handleKeyDown = (e: React.KeyboardEvent) => {
		if (value.trim() === '') {
			if (e.key === 'Enter') {
				e.preventDefault()
			}
		} else {
			if (e.key === 'Enter' && !e.shiftKey) {
				e.preventDefault()
				handleSubmit(onSubmit)()
			}
		}
	}

	// Сброс высоты текстового поля, чтобы затем установить его на scrollHeight
	useEffect(() => {
		if (textareaElement.current) {
			textareaElement.current.style.height = 'auto'
			textareaElement.current.style.height = `${textareaElement.current.scrollHeight}px`
		}
	}, [value, submitState])
	
	return (
		<footer className='flex flex-col gap-4 background-400'>
			<hr className='w-full border-background-400' />
			<form className='flex gap-4' ref={formElement} onSubmit={handleSubmit(onSubmit)}>
				<textarea 
					{...register('content', {required: true})}
					ref={textareaElement}
					tabIndex={0}
					rows={1}
					style={{ 
						height: 'auto', 
						maxHeight: '200px',
						overflowY: 'auto'
					}}

					className='w-full h-fit resize-none text-white placeholder:text-subtitle-gray font-bold bg-background-400 appearance-none py-3 px-4 rounded-xl border-subtitle-gray outline-none'
				
					placeholder='Type here...'
					onKeyDown={handleKeyDown} 
					onInput={(e) => {
						const target = e.target as HTMLTextAreaElement
						target.style.height = `${target.scrollHeight}px`
						handleChange(e)
					}}>
				</textarea>

				<button type='submit' className='bg-primary rounded-xl px-3	h-fit aspect-[1/1]'>
					<img className='w-6' src='../../../public/uil_message.svg' alt='' />
				</button>
			</form>
		</footer>
	)
}

export default Footer