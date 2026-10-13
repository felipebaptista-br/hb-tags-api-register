-- Seed das categorias do MVP (executar depois de schema.sql)

INSERT INTO category (slug, name, fields_schema) VALUES (
  'pet', 'Pets',
  $json${
  "version": 1,
  "fields": [
    {
      "key": "identity.display_name",
      "label": "Nome do pet",
      "type": "text",
      "required": true,
      "group": "identity",
      "visibility": "public",
      "maxLength": 40
    },
    {
      "key": "identity.photo_url",
      "label": "Foto",
      "type": "image",
      "required": true,
      "group": "identity",
      "visibility": "public"
    },
    {
      "key": "extra.species",
      "label": "Espécie",
      "type": "enum",
      "required": true,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "dog",
          "label": "Cachorro"
        },
        {
          "value": "cat",
          "label": "Gato"
        },
        {
          "value": "other",
          "label": "Outro"
        }
      ]
    },
    {
      "key": "extra.breed",
      "label": "Raça",
      "type": "text",
      "required": true,
      "group": "extra",
      "visibility": "public",
      "hint": "Aceita \"sem raça definida\""
    },
    {
      "key": "extra.sex",
      "label": "Sexo",
      "type": "enum",
      "required": true,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "male",
          "label": "Macho"
        },
        {
          "value": "female",
          "label": "Fêmea"
        }
      ]
    },
    {
      "key": "extra.neutered",
      "label": "Castrado",
      "type": "bool",
      "required": true,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.coat_color",
      "label": "Cor, pelagem e sinais particulares",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.birth_date",
      "label": "Data de nascimento",
      "type": "date",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.size",
      "label": "Porte",
      "type": "enum",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "small",
          "label": "Pequeno"
        },
        {
          "value": "medium",
          "label": "Médio"
        },
        {
          "value": "large",
          "label": "Grande"
        }
      ]
    },
    {
      "key": "extra.weight_kg",
      "label": "Peso (kg)",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.microchip",
      "label": "Número do microchip",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "restricted"
    },
    {
      "key": "extra.vaccinated_rabies",
      "label": "Antirrábica em dia",
      "type": "bool",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "health.allergies",
      "label": "Alergias",
      "type": "text_list",
      "required": false,
      "group": "health",
      "visibility": "public"
    },
    {
      "key": "health.critical_alerts",
      "label": "Alertas críticos",
      "type": "enum_multi",
      "required": false,
      "group": "health",
      "visibility": "public",
      "options": [
        {
          "value": "diabetes",
          "label": "Diabetes"
        },
        {
          "value": "epilepsy",
          "label": "Epilepsia"
        },
        {
          "value": "heart_condition",
          "label": "Problema cardíaco"
        }
      ]
    },
    {
      "key": "health.medications",
      "label": "Medicação contínua",
      "type": "medications",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.vet",
      "label": "Veterinário de referência",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "extra.temperament",
      "label": "Temperamento",
      "type": "enum",
      "required": true,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "docile",
          "label": "Dócil"
        },
        {
          "value": "shy",
          "label": "Medroso"
        },
        {
          "value": "may_bite",
          "label": "Pode morder"
        }
      ],
      "hint": "Segurança de quem encontrar o animal"
    },
    {
      "key": "contacts",
      "label": "Contatos do tutor",
      "type": "contacts",
      "required": true,
      "group": "contacts",
      "visibility": "public",
      "minItems": 1,
      "hint": "Primeiro nome e telefone ou WhatsApp"
    },
    {
      "key": "extra.tutor_city",
      "label": "Cidade do tutor",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.tutor_address",
      "label": "Endereço do tutor",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "private"
    },
    {
      "key": "extra.lost_message",
      "label": "Mensagem se estiver perdido",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "maxLength": 140,
      "hint": "Pode incluir recompensa"
    }
  ]
}$json$::jsonb
)
ON CONFLICT (slug) DO UPDATE
  SET name = EXCLUDED.name, fields_schema = EXCLUDED.fields_schema;

INSERT INTO category (slug, name, fields_schema) VALUES (
  'elderly', 'Idosos',
  $json${
  "version": 1,
  "fields": [
    {
      "key": "identity.display_name",
      "label": "Como gosta de ser chamado(a)",
      "type": "text",
      "required": true,
      "group": "identity",
      "visibility": "public",
      "maxLength": 60,
      "hint": "Prefira apelido ou primeiro nome"
    },
    {
      "key": "identity.photo_url",
      "label": "Foto",
      "type": "image",
      "required": false,
      "group": "identity",
      "visibility": "public"
    },
    {
      "key": "identity.age_range",
      "label": "Faixa etária",
      "type": "enum",
      "required": true,
      "group": "identity",
      "visibility": "public",
      "options": [
        {
          "value": "0-12",
          "label": "0 a 12 anos"
        },
        {
          "value": "13-17",
          "label": "13 a 17 anos"
        },
        {
          "value": "18-59",
          "label": "18 a 59 anos"
        },
        {
          "value": "60-69",
          "label": "60 a 69 anos"
        },
        {
          "value": "70-79",
          "label": "70 a 79 anos"
        },
        {
          "value": "80+",
          "label": "80 anos ou mais"
        }
      ]
    },
    {
      "key": "identity.birth_date",
      "label": "Data de nascimento",
      "type": "date",
      "required": false,
      "group": "identity",
      "visibility": "private"
    },
    {
      "key": "identity.language",
      "label": "Idioma",
      "type": "enum",
      "required": false,
      "group": "identity",
      "visibility": "public",
      "options": [
        {
          "value": "pt-BR",
          "label": "Português"
        },
        {
          "value": "en",
          "label": "Inglês"
        },
        {
          "value": "es",
          "label": "Espanhol"
        }
      ]
    },
    {
      "key": "health.blood_type",
      "label": "Tipo sanguíneo",
      "type": "enum",
      "required": false,
      "group": "health",
      "visibility": "public",
      "options": [
        {
          "value": "A+",
          "label": "A+"
        },
        {
          "value": "A-",
          "label": "A-"
        },
        {
          "value": "B+",
          "label": "B+"
        },
        {
          "value": "B-",
          "label": "B-"
        },
        {
          "value": "AB+",
          "label": "AB+"
        },
        {
          "value": "AB-",
          "label": "AB-"
        },
        {
          "value": "O+",
          "label": "O+"
        },
        {
          "value": "O-",
          "label": "O-"
        },
        {
          "value": "unknown",
          "label": "Não sei"
        }
      ]
    },
    {
      "key": "health.allergies",
      "label": "Alergias",
      "type": "text_list",
      "required": true,
      "group": "health",
      "visibility": "public",
      "hint": "Se não houver, informe \"nenhuma\""
    },
    {
      "key": "health.critical_alerts",
      "label": "Alertas críticos",
      "type": "enum_multi",
      "required": false,
      "group": "health",
      "visibility": "public",
      "options": [
        {
          "value": "uses_anticoagulant",
          "label": "Usa anticoagulante"
        },
        {
          "value": "uses_insulin",
          "label": "Usa insulina"
        },
        {
          "value": "has_pacemaker",
          "label": "Tem marca-passo"
        },
        {
          "value": "epilepsy",
          "label": "Epilepsia"
        }
      ]
    },
    {
      "key": "health.medications",
      "label": "Medicamentos em uso",
      "type": "medications",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.conditions",
      "label": "Condições de saúde",
      "type": "text_list",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.doctor",
      "label": "Médico ou serviço de referência",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.health_plan.provider",
      "label": "Plano de saúde (operadora)",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "private"
    },
    {
      "key": "health.health_plan.card_number",
      "label": "Plano de saúde (carteirinha)",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "private"
    },
    {
      "key": "contacts",
      "label": "Contatos de emergência",
      "type": "contacts",
      "required": true,
      "group": "contacts",
      "visibility": "public",
      "minItems": 1,
      "hint": "Cada contato será avisado de que foi cadastrado"
    },
    {
      "key": "instructions.note",
      "label": "Observação",
      "type": "textarea",
      "required": false,
      "group": "instructions",
      "visibility": "restricted",
      "maxLength": 280,
      "hint": "Não inclua dados desnecessários"
    },
    {
      "key": "extra.may_get_disoriented",
      "label": "Pode se desorientar ou se perder",
      "type": "bool",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.cognitive_diagnosis",
      "label": "Diagnóstico cognitivo",
      "type": "enum",
      "required": false,
      "group": "extra",
      "visibility": "restricted",
      "options": [
        {
          "value": "alzheimer",
          "label": "Alzheimer"
        },
        {
          "value": "dementia",
          "label": "Demência"
        },
        {
          "value": "other",
          "label": "Outro"
        }
      ],
      "hint": "Opcional, por decisão do responsável"
    },
    {
      "key": "extra.mobility_aids",
      "label": "Apoios de mobilidade",
      "type": "enum_multi",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "wheelchair",
          "label": "Cadeira de rodas"
        },
        {
          "value": "walker",
          "label": "Andador"
        },
        {
          "value": "cane",
          "label": "Bengala"
        }
      ]
    },
    {
      "key": "extra.sensory_aids",
      "label": "Óculos, aparelho auditivo ou prótese",
      "type": "enum_multi",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "glasses",
          "label": "Óculos"
        },
        {
          "value": "hearing_aid",
          "label": "Aparelho auditivo"
        },
        {
          "value": "prosthesis",
          "label": "Prótese"
        }
      ]
    },
    {
      "key": "extra.communication_tips",
      "label": "Como se comunicar",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "restricted",
      "maxLength": 140
    },
    {
      "key": "extra.home_city",
      "label": "Cidade e bairro",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "restricted"
    },
    {
      "key": "extra.home_address",
      "label": "Endereço completo",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "private"
    }
  ]
}$json$::jsonb
)
ON CONFLICT (slug) DO UPDATE
  SET name = EXCLUDED.name, fields_schema = EXCLUDED.fields_schema;

INSERT INTO category (slug, name, fields_schema) VALUES (
  'neurodivergent', 'Neurodivergentes',
  $json${
  "version": 1,
  "fields": [
    {
      "key": "identity.display_name",
      "label": "Como gosta de ser chamado(a)",
      "type": "text",
      "required": true,
      "group": "identity",
      "visibility": "public",
      "maxLength": 60,
      "hint": "Prefira apelido ou primeiro nome"
    },
    {
      "key": "identity.photo_url",
      "label": "Foto",
      "type": "image",
      "required": false,
      "group": "identity",
      "visibility": "private"
    },
    {
      "key": "identity.age_range",
      "label": "Faixa etária",
      "type": "enum",
      "required": true,
      "group": "identity",
      "visibility": "public",
      "options": [
        {
          "value": "0-12",
          "label": "0 a 12 anos"
        },
        {
          "value": "13-17",
          "label": "13 a 17 anos"
        },
        {
          "value": "18-59",
          "label": "18 a 59 anos"
        },
        {
          "value": "60-69",
          "label": "60 a 69 anos"
        },
        {
          "value": "70-79",
          "label": "70 a 79 anos"
        },
        {
          "value": "80+",
          "label": "80 anos ou mais"
        }
      ]
    },
    {
      "key": "identity.birth_date",
      "label": "Data de nascimento",
      "type": "date",
      "required": false,
      "group": "identity",
      "visibility": "private"
    },
    {
      "key": "identity.language",
      "label": "Idioma",
      "type": "enum",
      "required": false,
      "group": "identity",
      "visibility": "public",
      "options": [
        {
          "value": "pt-BR",
          "label": "Português"
        },
        {
          "value": "en",
          "label": "Inglês"
        },
        {
          "value": "es",
          "label": "Espanhol"
        }
      ]
    },
    {
      "key": "health.blood_type",
      "label": "Tipo sanguíneo",
      "type": "enum",
      "required": false,
      "group": "health",
      "visibility": "public",
      "options": [
        {
          "value": "A+",
          "label": "A+"
        },
        {
          "value": "A-",
          "label": "A-"
        },
        {
          "value": "B+",
          "label": "B+"
        },
        {
          "value": "B-",
          "label": "B-"
        },
        {
          "value": "AB+",
          "label": "AB+"
        },
        {
          "value": "AB-",
          "label": "AB-"
        },
        {
          "value": "O+",
          "label": "O+"
        },
        {
          "value": "O-",
          "label": "O-"
        },
        {
          "value": "unknown",
          "label": "Não sei"
        }
      ]
    },
    {
      "key": "health.allergies",
      "label": "Alergias",
      "type": "text_list",
      "required": true,
      "group": "health",
      "visibility": "public",
      "hint": "Se não houver, informe \"nenhuma\""
    },
    {
      "key": "health.critical_alerts",
      "label": "Alertas críticos",
      "type": "enum_multi",
      "required": false,
      "group": "health",
      "visibility": "public",
      "options": [
        {
          "value": "uses_anticoagulant",
          "label": "Usa anticoagulante"
        },
        {
          "value": "uses_insulin",
          "label": "Usa insulina"
        },
        {
          "value": "has_pacemaker",
          "label": "Tem marca-passo"
        },
        {
          "value": "epilepsy",
          "label": "Epilepsia"
        }
      ]
    },
    {
      "key": "health.medications",
      "label": "Medicamentos em uso",
      "type": "medications",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.conditions",
      "label": "Condições de saúde",
      "type": "text_list",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.doctor",
      "label": "Médico ou serviço de referência",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "restricted"
    },
    {
      "key": "health.health_plan.provider",
      "label": "Plano de saúde (operadora)",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "private"
    },
    {
      "key": "health.health_plan.card_number",
      "label": "Plano de saúde (carteirinha)",
      "type": "text",
      "required": false,
      "group": "health",
      "visibility": "private"
    },
    {
      "key": "contacts",
      "label": "Contatos de emergência",
      "type": "contacts",
      "required": true,
      "group": "contacts",
      "visibility": "public",
      "minItems": 1,
      "hint": "Cada contato será avisado de que foi cadastrado"
    },
    {
      "key": "instructions.note",
      "label": "Observação",
      "type": "textarea",
      "required": false,
      "group": "instructions",
      "visibility": "restricted",
      "maxLength": 280,
      "hint": "Não inclua dados desnecessários"
    },
    {
      "key": "extra.communication_mode",
      "label": "Forma de comunicação",
      "type": "enum",
      "required": true,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "verbal",
          "label": "Verbal"
        },
        {
          "value": "partial",
          "label": "Verbal parcial"
        },
        {
          "value": "non_verbal",
          "label": "Não verbal"
        },
        {
          "value": "uses_aids",
          "label": "Usa pranchas ou gestos"
        }
      ]
    },
    {
      "key": "extra.responds_to_name",
      "label": "Responde ao nome",
      "type": "bool",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.sensory_triggers",
      "label": "Gatilhos sensoriais",
      "type": "enum_multi",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "noise",
          "label": "Barulho"
        },
        {
          "value": "light",
          "label": "Luz"
        },
        {
          "value": "touch",
          "label": "Toque"
        },
        {
          "value": "crowd",
          "label": "Multidão"
        },
        {
          "value": "smell",
          "label": "Cheiros"
        }
      ]
    },
    {
      "key": "extra.crisis_signs",
      "label": "Sinais de crise",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "restricted",
      "maxLength": 140
    },
    {
      "key": "instructions.what_helps",
      "label": "O que ajuda a acalmar",
      "type": "text",
      "required": false,
      "group": "instructions",
      "visibility": "public",
      "maxLength": 140
    },
    {
      "key": "instructions.what_to_avoid",
      "label": "O que evitar",
      "type": "text",
      "required": true,
      "group": "instructions",
      "visibility": "public",
      "maxLength": 140,
      "hint": "Campo mais importante para quem for ajudar"
    },
    {
      "key": "extra.elopement_risk",
      "label": "Risco de fuga",
      "type": "bool",
      "required": false,
      "group": "extra",
      "visibility": "public"
    },
    {
      "key": "extra.danger_awareness",
      "label": "Noção de perigo (trânsito, água)",
      "type": "enum",
      "required": false,
      "group": "extra",
      "visibility": "restricted",
      "options": [
        {
          "value": "none",
          "label": "Nenhuma"
        },
        {
          "value": "partial",
          "label": "Parcial"
        },
        {
          "value": "good",
          "label": "Boa"
        }
      ]
    },
    {
      "key": "extra.clinical_alert",
      "label": "Condição clínica associada",
      "type": "enum_multi",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "options": [
        {
          "value": "seizures",
          "label": "Convulsões"
        }
      ]
    },
    {
      "key": "extra.diagnosis",
      "label": "Diagnóstico",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "private",
      "hint": "Opcional; o responsável pode torná-lo restrito"
    },
    {
      "key": "extra.approach_card",
      "label": "Cartão de abordagem",
      "type": "textarea",
      "required": false,
      "group": "extra",
      "visibility": "public",
      "maxLength": 200,
      "hint": "Até 3 linhas para socorristas e policiais"
    },
    {
      "key": "extra.therapist",
      "label": "Terapeuta de referência",
      "type": "text",
      "required": false,
      "group": "extra",
      "visibility": "private"
    }
  ]
}$json$::jsonb
)
ON CONFLICT (slug) DO UPDATE
  SET name = EXCLUDED.name, fields_schema = EXCLUDED.fields_schema;
