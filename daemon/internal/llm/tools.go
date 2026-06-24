package llm

// Tools returns the v1 internal tool set exposed to the model (docs §6.4).
// Argument names are Portuguese to match the domain vocabulary the model sees.
func Tools() []Tool {
	str := func(desc string) map[string]any { return map[string]any{"type": "string", "description": desc} }
	return []Tool{
		{
			Name:        "criar_evento",
			Description: "Cria um evento no calendário. Use rrule para recorrência (iCalendar) ou data para evento único.",
			Parameters: object(map[string]any{
				"titulo":   str("Título do evento"),
				"contexto": str("Empresa/contexto (ex.: Cogna, Visa)"),
				"inicio":   str("Horário de início HH:MM"),
				"fim":      str("Horário de fim HH:MM"),
				"rrule":    str("RRULE iCalendar, ex.: FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"),
				"data":     str("Data YYYY-MM-DD para evento único"),
			}, "titulo", "inicio"),
		},
		{
			Name:        "criar_tarefa",
			Description: "Cria uma tarefa/lembrete.",
			Parameters: object(map[string]any{
				"titulo":   str("Descrição da tarefa"),
				"contexto": str("Empresa/contexto"),
			}, "titulo"),
		},
		{
			Name:        "consultar_agenda",
			Description: "Resume a agenda e tarefas do dia.",
			Parameters:  object(map[string]any{}),
		},
		{
			Name:        "consultar_vps",
			Description: "Consulta o estado dos servidores/VPS.",
			Parameters:  object(map[string]any{}),
		},
		{
			Name:        "registrar_nota",
			Description: "Registra uma nota livre na memória.",
			Parameters: object(map[string]any{
				"texto": str("Conteúdo da nota"),
			}, "texto"),
		},
	}
}

func object(props map[string]any, required ...string) map[string]any {
	m := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		m["required"] = required
	}
	return m
}
