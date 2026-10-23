
      SELECT
          s.id,
          s.video_url,
          svc.service_key AS service,
          p.name AS provider
      FROM public.subtitles AS s
      LEFT JOIN public.subtitle_provider_services AS svc
          ON svc.id = s.service_id
      LEFT JOIN public.subtitle_providers AS p
          ON p.id = s.provider_id
      WHERE s.id = ANY($1);